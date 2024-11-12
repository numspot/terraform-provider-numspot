package snapshot

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewSnapshotResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_snapshot"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = SnapshotResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan SnapshotModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
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

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state SnapshotModel
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

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan SnapshotModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	snapshotID := state.Id.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

	var numspotSnapshot *numspot.Snapshot
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

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state SnapshotModel
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

func deserializeCreateSnapshot(tf SnapshotModel) numspot.CreateSnapshotJSONRequestBody {
	return numspot.CreateSnapshotJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		SourceRegionName: tf.SourceRegionName.ValueStringPointer(),
		SourceSnapshotId: tf.SourceSnapshotId.ValueStringPointer(),
		VolumeId:         tf.VolumeId.ValueStringPointer(),
	}
}

func serializeSnapshot(ctx context.Context, http *numspot.Snapshot, model SnapshotModel, diags *diag.Diagnostics) *SnapshotModel {
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

	snapshot := SnapshotModel{
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
