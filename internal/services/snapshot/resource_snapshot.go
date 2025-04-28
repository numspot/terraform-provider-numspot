package snapshot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
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

type snapshotResource struct {
	provider *client.NumSpotSDK
}

func NewSnapshotResource() resource.Resource {
	return &snapshotResource{}
}

func (r *snapshotResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *snapshotResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *snapshotResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_snapshot"
}

func (r *snapshotResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_snapshot.SnapshotResourceSchema(ctx)
}

func (r *snapshotResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := snapshotTags(ctx, plan.Tags)
	body := deserializeCreateSnapshot(plan)
	if response.Diagnostics.HasError() {
		return
	}

	snapshot, err := core.CreateSnapshot(ctx, r.provider, tagsValue, body)
	if err != nil {
		response.Diagnostics.AddError("unable to create snapshot", err.Error())
		return
	}

	tf := serializeSnapshot(ctx, snapshot, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *snapshotResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	snapshotID := state.Id.ValueString()

	snapshot, err := core.ReadSnapshot(ctx, r.provider, snapshotID)
	if err != nil {
		response.Diagnostics.AddError("unable to read snapshot", err.Error())
		return
	}

	tf := serializeSnapshot(ctx, snapshot, state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *snapshotResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resource_snapshot.SnapshotModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	snapshotID := state.Id.ValueString()
	planTags := snapshotTags(ctx, plan.Tags)
	stateTags := snapshotTags(ctx, state.Tags)

	var numspotSnapshot *api.Snapshot
	var err error
	if !state.Tags.Equal(plan.Tags) {
		numspotSnapshot, err = core.UpdateSnapshotTags(ctx, r.provider, stateTags, planTags, snapshotID)
		if err != nil {
			response.Diagnostics.AddError("unable to update snapshot tags", err.Error())
			return
		}
	}

	tf := serializeSnapshot(ctx, numspotSnapshot, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *snapshotResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	snapshotID := state.Id.ValueString()
	err := core.DeleteSnapshot(ctx, r.provider, snapshotID)
	if err != nil {
		response.Diagnostics.AddError("unable to delete snapshot", err.Error())
		return
	}
}

func deserializeCreateSnapshot(tf resource_snapshot.SnapshotModel) api.CreateSnapshotJSONRequestBody {
	return api.CreateSnapshotJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		SourceRegionName: tf.SourceRegionName.ValueStringPointer(),
		SourceSnapshotId: tf.SourceSnapshotId.ValueStringPointer(),
		VolumeId:         tf.VolumeId.ValueStringPointer(),
	}
}

func serializeSnapshot(ctx context.Context, http *api.Snapshot, model resource_snapshot.SnapshotModel, diags *diag.Diagnostics) *resource_snapshot.SnapshotModel {
	var (
		tagsTf          types.List
		creationDateStr *string
	)

	if http.CreationDate != nil {
		tmp := (*http.CreationDate).String()
		creationDateStr = &tmp
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	snapshot := resource_snapshot.SnapshotModel{
		CreationDate: types.StringPointerValue(creationDateStr),
		Description:  types.StringPointerValue(http.Description),
		Id:           types.StringPointerValue(http.Id),
		Progress:     utils.FromIntPtrToTfInt64(http.Progress),
		State:        types.StringPointerValue(http.State),
		VolumeId:     types.StringPointerValue(http.VolumeId),
		VolumeSize:   utils.FromIntPtrToTfInt64(http.VolumeSize),
		Tags:         tagsTf,
	}

	if !model.SourceRegionName.IsUnknown() {
		snapshot.SourceRegionName = model.SourceRegionName
	} else {
		snapshot.SourceRegionName = types.StringNull()
	}

	if !model.SourceSnapshotId.IsUnknown() {
		snapshot.SourceSnapshotId = model.SourceSnapshotId
	} else {
		snapshot.SourceSnapshotId = types.StringNull()
	}

	return &snapshot
}

func snapshotTags(ctx context.Context, tags types.List) []api.ResourceTag {
	tfTags := make([]resource_snapshot.TagsValue, 0, len(tags.Elements()))
	tags.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
