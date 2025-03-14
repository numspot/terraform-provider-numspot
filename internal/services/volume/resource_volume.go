package volume

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/core"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewVolumeResource() resource.Resource {
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
	response.TypeName = request.ProviderTypeName + "_volume"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VolumeResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan VolumeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vmID := plan.LinkVM.VmID.ValueString()
	deviceName := plan.LinkVM.DeviceName.ValueString()
	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)

	numSpotVolume, err := core.CreateVolume(ctx, r.provider, deserializeCreateNumSpotVolume(plan), tagsValue, vmID, deviceName)
	if err != nil {
		response.Diagnostics.AddError("unable to create volume", err.Error())
		return
	}

	state := serializeNumSpotVolume(ctx, numSpotVolume, &response.Diagnostics, plan.ReplaceVolumeOnDownsize)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state VolumeModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	volumeID := state.Id.ValueString()

	numSpotVolume, err := core.ReadVolume(ctx, r.provider, volumeID)
	if err != nil {
		response.Diagnostics.AddError("unable to read volume", err.Error())
		return
	}

	tf := serializeNumSpotVolume(ctx, numSpotVolume, &response.Diagnostics, state.ReplaceVolumeOnDownsize)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err           error
		numSpotVolume *numspot.Volume
		state, plan   VolumeModel
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if plan.Size.ValueInt64() < state.Size.ValueInt64() {
		response.Diagnostics.AddError("volume downsize", fmt.Sprintf("Trying to update volume size from %v to %v. It is not possible to downsize a volume in an update. "+
			"To force the replace of volume, set attribute 'replace_volume_on_downsize' to true. Note : All data on volume will be lost.", state.Size.ValueInt64(), plan.Size.ValueInt64()))
		return
	}

	volumeID := state.Id.ValueString()
	stateVMID := state.LinkVM.VmID.ValueString()
	planVMID := plan.LinkVM.VmID.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)
	newDeviceName := plan.LinkVM.DeviceName.ValueString()

	if !plan.Size.Equal(state.Size) || !plan.Type.Equal(state.Type) || (!utils.IsTfValueNull(plan.Iops) && !plan.Iops.Equal(state.Iops)) {
		numSpotVolume, err = core.UpdateVolumeAttributes(ctx, r.provider, deserializeUpdateNumspotVolume(plan), volumeID, stateVMID)
		if err != nil {
			response.Diagnostics.AddError("unable to update volume attributes", err.Error())
			return
		}
	}

	if !plan.LinkVM.VmID.Equal(state.LinkVM.VmID) || !plan.LinkVM.DeviceName.Equal(state.LinkVM.DeviceName) {
		numSpotVolume, err = core.UpdateVolumeLink(ctx, r.provider, volumeID, stateVMID, planVMID, newDeviceName)
		if err != nil {
			response.Diagnostics.AddError("unable to update volume link", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numSpotVolume, err = core.UpdateVolumeTags(ctx, r.provider, volumeID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update volume tags", err.Error())
			return
		}
	}

	newState := serializeNumSpotVolume(ctx, numSpotVolume, &response.Diagnostics, plan.ReplaceVolumeOnDownsize)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state VolumeModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVolume(ctx, r.provider, state.Id.ValueString(), state.LinkVM.VmID.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete volume", err.Error())
		return
	}
}

func serializeNumSpotVolume(ctx context.Context, http *numspot.Volume, diags *diag.Diagnostics, ReplaceVolumeOnDownsize basetypes.BoolValue) VolumeModel {
	var (
		volumes = types.ListNull(LinkedVolumesValue{}.Type(ctx))
		tagsTf  types.List
		linkVm  LinkVMValue
	)

	if http.LinkedVolumes != nil {
		volumes = utils.GenericListToTfListValue(
			ctx,
			serializeLinkedVolumes,
			*http.LinkedVolumes,
			diags,
		)

		nbLinkedVolumes := len(*http.LinkedVolumes)
		if nbLinkedVolumes > 0 {
			var diagnostics diag.Diagnostics
			linkVm, diagnostics = NewLinkVMValue(LinkVMValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"device_name": types.StringPointerValue((*http.LinkedVolumes)[0].DeviceName),
					"vm_id":       types.StringPointerValue((*http.LinkedVolumes)[0].VmId),
				})
			diags.Append(diagnostics...)
		}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return VolumeModel{
		CreationDate:            types.StringValue(http.CreationDate.String()),
		Id:                      types.StringPointerValue(http.Id),
		Iops:                    utils.FromIntPtrToTfInt64(http.Iops),
		Size:                    utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:              types.StringPointerValue(http.SnapshotId),
		State:                   types.StringPointerValue(http.State),
		AvailabilityZoneName:    types.StringPointerValue(http.AvailabilityZoneName),
		Type:                    types.StringPointerValue(http.Type),
		LinkedVolumes:           volumes,
		Tags:                    tagsTf,
		LinkVM:                  linkVm,
		ReplaceVolumeOnDownsize: ReplaceVolumeOnDownsize,
	}
}

func serializeLinkedVolumes(ctx context.Context, http numspot.LinkedVolume, diags *diag.Diagnostics) LinkedVolumesValue {
	value, diagnostics := NewLinkedVolumesValue(
		LinkedVolumesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_name":           types.StringPointerValue(http.DeviceName),
			"state":                 types.StringPointerValue(http.State),
			"vm_id":                 types.StringPointerValue(http.VmId),
			"id":                    types.StringPointerValue(http.Id),
		})
	diags.Append(diagnostics...)
	return value
}

func deserializeCreateNumSpotVolume(tf VolumeModel) numspot.CreateVolumeJSONRequestBody {
	var (
		httpIOPS   *int
		snapshotId *string
	)
	if !tf.Iops.IsUnknown() && !tf.Iops.IsNull() {
		httpIOPS = utils.FromTfInt64ToIntPtr(tf.Iops)
	}
	if !tf.SnapshotId.IsUnknown() {
		snapshotId = tf.SnapshotId.ValueStringPointer()
	}

	return numspot.CreateVolumeJSONRequestBody{
		Iops:                 httpIOPS,
		Size:                 utils.FromTfInt64ToIntPtr(tf.Size),
		SnapshotId:           snapshotId,
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueString(),
		Type:                 tf.Type.ValueStringPointer(),
	}
}

func deserializeUpdateNumspotVolume(tf VolumeModel) numspot.UpdateVolumeJSONRequestBody {
	var httpIOPS *int
	if !tf.Iops.IsUnknown() && !tf.Iops.IsNull() {
		httpIOPS = utils.FromTfInt64ToIntPtr(tf.Iops)
	}

	return numspot.UpdateVolumeJSONRequestBody{
		Iops:       httpIOPS,
		Size:       utils.FromTfInt64ToIntPtr(tf.Size),
		VolumeType: tf.Type.ValueStringPointer(),
	}
}
