package snapshot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/snapshot/resource_snapshot"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &snapshotResource{}
	_ resource.ResourceWithConfigure   = &snapshotResource{}
	_ resource.ResourceWithImportState = &snapshotResource{}
)

// snapshotResource is the Terraform resource implementation for numspot_snapshot.
type snapshotResource struct {
	services.BaseResource
}

// NewSnapshotResource returns a new snapshotResource instance.
func NewSnapshotResource() resource.Resource {
	return &snapshotResource{}
}

// Metadata sets the resource type name.
func (r *snapshotResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_snapshot"
}

// Schema defines the schema for the resource.
func (r *snapshotResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_snapshot.SnapshotResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *snapshotResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := snapshotTags(ctx, plan.Tags)
	numSpotSnapshot, err := core.CreateSnapshot(ctx, r.Provider, deserializeCreateSnapshotRequest(&plan), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create snapshot", err.Error())
		return
	}

	state := serializeSnapshot(ctx, numSpotSnapshot, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	state.SourceRegionName = utils.PreserveFromState(plan.SourceRegionName)
	state.SourceSnapshotId = utils.PreserveFromState(plan.SourceSnapshotId)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *snapshotResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotSnapshot, err := core.ReadSnapshot(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read snapshot", err.Error())
		return
	}

	newState := serializeSnapshot(ctx, numSpotSnapshot, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.SourceRegionName = state.SourceRegionName
	newState.SourceSnapshotId = state.SourceSnapshotId

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *snapshotResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err             error
		state, plan     resource_snapshot.SnapshotModel
		numSpotSnapshot *api.Snapshot
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	snapshotID := state.Id.ValueString()
	planTags := snapshotTags(ctx, plan.Tags)
	stateTags := snapshotTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotSnapshot, err = core.UpdateSnapshotTags(ctx, r.Provider, snapshotID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update snapshot tags", err.Error())
			return
		}
	}

	newState := serializeSnapshot(ctx, numSpotSnapshot, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.SourceRegionName = state.SourceRegionName
	newState.SourceSnapshotId = state.SourceSnapshotId

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *snapshotResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteSnapshot(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete snapshot", err.Error())
		return
	}
}

func serializeSnapshot(ctx context.Context, http *api.Snapshot, diags *diag.Diagnostics) resource_snapshot.SnapshotModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil Snapshot response", "the API returned an empty Snapshot")
		return resource_snapshot.SnapshotModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_snapshot.SnapshotModel{
		Access:       serializeAccess(ctx, http.Access, diags),
		CreationDate: utils.TimePtrToTfString(http.CreationDate),
		Description:  types.StringPointerValue(http.Description),
		Id:           types.StringPointerValue(http.Id),
		Progress:     utils.FromIntPtrToTfInt64(http.Progress),
		State:        types.StringPointerValue(http.State),
		VolumeId:     types.StringPointerValue(http.VolumeId),
		VolumeSize:   utils.FromIntPtrToTfInt64(http.VolumeSize),
		Tags:         tagsTf,
	}
}

func deserializeCreateSnapshotRequest(tf *resource_snapshot.SnapshotModel) *api.CreateSnapshotJSONRequestBody {
	return &api.CreateSnapshotJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		SourceRegionName: tf.SourceRegionName.ValueStringPointer(),
		SourceSnapshotId: tf.SourceSnapshotId.ValueStringPointer(),
		VolumeId:         tf.VolumeId.ValueStringPointer(),
	}
}

func snapshotTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_snapshot.TagsValue, 0, len(tagsSet.Elements()))
	tagsSet.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}

func serializeAccess(ctx context.Context, item *api.Access, diags *diag.Diagnostics) resource_snapshot.AccessValue {
	if item == nil {
		return resource_snapshot.NewAccessValueNull()
	}
	value, d := resource_snapshot.NewAccessValue(resource_snapshot.AccessValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"is_public": types.BoolPointerValue(item.IsPublic),
	})
	diags.Append(d...)
	return value
}
