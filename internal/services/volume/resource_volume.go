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
	var initialPlan VolumeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &initialPlan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vmID := initialPlan.LinkVM.VmID.ValueString()
	deviceName := initialPlan.LinkVM.DeviceName.ValueString()

	numSpotVolume, err := core.CreateVolume(ctx, r.provider, deserializeCreateVolume(initialPlan), vmID, deviceName)
	if err != nil {
		response.Diagnostics.AddError("unable to create volume", err.Error())
		return
	}

	newPlan, diags := serializeNumSpotVolume(ctx, numSpotVolume)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	// uhh
	newPlan.Tags = initialPlan.Tags

	//createStateConf := &retry.StateChangeConf{
	//	Pending: []string{"attaching"},
	//	Target:  []string{"attached"},
	//	Refresh: func() (interface{}, string, error) {
	//		readRes, err := r.provider.GetNumspotClient().ReadVolumesByIdWithResponse(ctx, r.provider.GetSpaceID(), createdId)
	//		if err != nil {
	//			return nil, "", fmt.Errorf("failed to read volume : %v", err)
	//		}
	//
	//		if len(*readRes.JSON200.LinkedVolumes) > 0 {
	//			linkState := (*readRes.JSON200.LinkedVolumes)[0].State
	//			return readRes.JSON200, *linkState, nil
	//		}
	//
	//		return nil, "", fmt.Errorf("Volume not linked to any VM : %v", err)
	//	},
	//	Timeout: utils.TfRequestRetryTimeout,
	//	Delay:   utils.TfRequestRetryDelay,
	//}
	//read, err = createStateConf.WaitForStateContext(ctx)
	//if err != nil {
	//	response.Diagnostics.AddError("Failed to create volume", fmt.Sprintf("Error waiting for volume (%s) to be linked: %s", *res.JSON201.Id, err))
	//	return
	//}

	if len(initialPlan.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(
			ctx,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			&response.Diagnostics,
			newPlan.Id.ValueString(),
			initialPlan.Tags,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newPlan)...)
}

func (r *VolumeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var initialState VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &initialState)...)

	res, err := r.provider.GetNumspotClient().ReadVolumesByIdWithResponse(ctx, r.provider.GetSpaceID(), initialState.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("", "")
		return
	}
	if res.JSON200 == nil {
		response.Diagnostics.AddError("", "")
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
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	volumeID := plan.Id.ValueString()
	oldVM := plan.LinkVM.VmID.ValueString()
	newVM := state.LinkVM.VmID.ValueString()
	deviceName := state.LinkVM.DeviceName.ValueString()

	if (!utils.IsTfValueNull(plan.Size) && !plan.Size.Equal(state.Size)) ||
		(!utils.IsTfValueNull(plan.Type) && !plan.Type.Equal(state.Type)) ||
		(!utils.IsTfValueNull(plan.Iops) && !plan.Iops.Equal(state.Iops)) {
		body := ValueFromTfToUpdaterequest(&plan)
		numSpotVolume, err = core.UpdateVolumeAttributes(ctx, r.provider, volumeID, body, oldVM, "")
		if err != nil {
			response.Diagnostics.AddError("unable to update volume", err.Error())
		}
	}

	if !utils.IsTfValueNull(plan.LinkVM.VmID) && !plan.Size.Equal(state.LinkVM.VmID) {
		numSpotVolume, err = core.UpdateVolumeLink(ctx, r.provider, volumeID, oldVM, newVM, deviceName)
	}

	newState, diags := serializeNumSpotVolume(ctx, numSpotVolume)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *VolumeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//var data VolumeModel
	//
	//response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	//
	//if linkVmConfigured(data) {
	//	// Stop linked VM before unlinking Volume
	//	diags, stateVmExists := r.stopVm(ctx, data.LinkVM.VmID.ValueString())
	//
	//	if diags.HasError() {
	//		response.Diagnostics.Append(diags...)
	//		return
	//	}
	//
	//	if stateVmExists {
	//		// Unlink VM
	//		utils.ExecuteRequest(func() (*numspot.UnlinkVolumeResponse, error) {
	//			return r.provider.GetNumspotClient().UnlinkVolumeWithResponse(
	//				ctx,
	//				r.provider.GetSpaceID(),
	//				data.Id.ValueString(),
	//				VolumeFromTfToUnlinkRequest(&data),
	//			)
	//		}, http.StatusNoContent, &response.Diagnostics)
	//		if response.Diagnostics.HasError() {
	//			return
	//		}
	//		diags = vm.StartVm(ctx, r.provider, data.LinkVM.VmID.ValueString())
	//		response.Diagnostics.Append(diags...)
	//
	//		if response.Diagnostics.HasError() {
	//			return
	//		}
	//
	//	}
	//}
	//
	//err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteVolumeWithResponse)
	//if err != nil {
	//	response.Diagnostics.AddError("Failed to delete Volume", err.Error())
	//	return
	//}
}

func deserializeCreateVolume(tf VolumeModel) numspot.CreateVolumeJSONRequestBody {
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

		nbLinkedVolumes := len((*http.LinkedVolumes))
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
