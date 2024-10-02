package volume

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VolumeResource{}
	_ resource.ResourceWithConfigure   = &VolumeResource{}
	_ resource.ResourceWithImportState = &VolumeResource{}
)

type VolumeResource struct {
	provider services.IProvider
}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

func (r *VolumeResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *VolumeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VolumeResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VolumeResourceSchema(ctx)
}

func (r *VolumeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
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

	state, diags := serializeNumSpotVolume(ctx, numSpotVolume)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *VolumeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := r.provider.GetNumspotClient().ReadVolumesByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read volume", err.Error())
		return
	}

	if res.JSON200 == nil {
		response.Diagnostics.AddError("unable to read volume", "empty response")
		return
	}

	newState, diags := serializeNumSpotVolume(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *VolumeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err           error
		numSpotVolume *numspot.Volume
		state, plan   VolumeModel
		diags         diag.Diagnostics
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
		response.Diagnostics.AddError("volume downscale", "you can't downscale a volume")
		return
	}

	volumeID := state.Id.ValueString()
	stateVMID := state.LinkVM.VmID.ValueString()
	planVMID := plan.LinkVM.VmID.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)
	newDeviceName := plan.LinkVM.DeviceName.ValueString()

	if !plan.Size.Equal(state.Size) || !plan.Type.Equal(state.Type) || (!utils.IsTfValueNull(plan.Iops) && !plan.Iops.Equal(state.Iops)) {
		numSpotVolume, err = core.UpdateVolumeAttributes(ctx, r.provider, deserializeUpdateNumspotVolume(plan), volumeID, stateVMID, planVMID)
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

	newState, diags := serializeNumSpotVolume(ctx, numSpotVolume)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *VolumeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state VolumeModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVolume(ctx, r.provider, state.Id.ValueString(), state.LinkVM.VmID.ValueString(), ""); err != nil {
		response.Diagnostics.AddError("failed to delete volume", err.Error())
		return
	}
}

func serializeNumSpotVolume(ctx context.Context, http *numspot.Volume) (VolumeModel, diag.Diagnostics) {
	var (
		volumes = types.ListNull(LinkedVolumesValue{}.Type(ctx))
		tagsTf  types.List
		diags   diag.Diagnostics
		linkVm  LinkVMValue
	)

	if http.LinkedVolumes != nil {
		volumes, diags = utils.GenericListToTfListValue(
			ctx,
			LinkedVolumesValue{},
			serializeLinkedVolumes,
			*http.LinkedVolumes,
		)

		if diags.HasError() {
			return VolumeModel{}, diags
		}

		nbLinkedVolumes := len(*http.LinkedVolumes)
		if nbLinkedVolumes > 0 {
			linkVm, diags = NewLinkVMValue(LinkVMValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"device_name": types.StringPointerValue((*http.LinkedVolumes)[0].DeviceName),
					"vm_id":       types.StringPointerValue((*http.LinkedVolumes)[0].VmId),
				})
			if diags.HasError() {
				return VolumeModel{}, diags
			}

		}

		if diags.HasError() {
			return VolumeModel{}, diags
		}
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return VolumeModel{}, diags
		}
	}

	return VolumeModel{
		CreationDate:         types.StringValue(http.CreationDate.String()),
		Id:                   types.StringPointerValue(http.Id),
		Iops:                 utils.FromIntPtrToTfInt64(http.Iops),
		Size:                 utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:           types.StringPointerValue(http.SnapshotId),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Type:                 types.StringPointerValue(http.Type),
		LinkedVolumes:        volumes,
		Tags:                 tagsTf,
		LinkVM:               linkVm,
	}, diags
}

func serializeLinkedVolumes(ctx context.Context, http numspot.LinkedVolume) (LinkedVolumesValue, diag.Diagnostics) {
	return NewLinkedVolumesValue(
		LinkedVolumesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_name":           types.StringPointerValue(http.DeviceName),
			"state":                 types.StringPointerValue(http.State),
			"vm_id":                 types.StringPointerValue(http.VmId),
			"id":                    types.StringPointerValue(http.Id),
		})
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
