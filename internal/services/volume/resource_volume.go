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
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/services/volume/resource_volume"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &volumeResource{}
	_ resource.ResourceWithConfigure   = &volumeResource{}
	_ resource.ResourceWithImportState = &volumeResource{}
)

// Package volume provides the implementation of the Volume resource
// for the NumSpot provider. It handles the creation, reading, updating, and
// deletion of volumes, along with their attachment to VMs and management
// of volume attributes like size, type, and IOPS.

// volumeResource represents the Volume resource implementation.
// It implements the Terraform resource.Resource interface and provides
// methods for managing the lifecycle of volumes.
type volumeResource struct {
	provider *client.NumSpotSDK
}

// NewVolumeResource creates a new instance of the Volume resource.
func NewVolumeResource() resource.Resource {
	return &volumeResource{}
}

// Configure sets up the provider client for the Volume resource.
func (r *volumeResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

// ImportState handles importing an existing Volume into Terraform state.
func (r *volumeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// Metadata sets the resource type name for the Volume resource.
func (r *volumeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_volume"
}

// Schema returns the Terraform schema for the Volume resource.
func (r *volumeResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_volume.VolumeResourceSchema(ctx)
}

// Create handles the creation of a new Volume.
// It deserializes the plan into a NumSpot Volume creation request,
// creates the Volume, and updates the state with the created Volume's details.
func (r *volumeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_volume.VolumeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vmID := plan.LinkVm.VmId.ValueString()
	deviceName := plan.LinkVm.DeviceName.ValueString()
	tagsValue := volumeTags(ctx, plan.Tags)

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

// Read retrieves the current state of a Volume and updates the Terraform state.
func (r *volumeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_volume.VolumeModel

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

// Update handles updates to an existing Volume.
// It supports updating various Volume attributes including:
// - Size (with downsize protection)
// - Type
// - IOPS
// - VM attachment
// - Tags
func (r *volumeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err           error
		numSpotVolume *api.Volume
		state, plan   resource_volume.VolumeModel
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
	stateVMID := state.LinkVm.VmId.ValueString()
	planVMID := plan.LinkVm.VmId.ValueString()
	planTags := volumeTags(ctx, plan.Tags)
	stateTags := volumeTags(ctx, state.Tags)
	newDeviceName := plan.LinkVm.DeviceName.ValueString()

	if !plan.Size.Equal(state.Size) || !plan.Type.Equal(state.Type) || (!utils.IsTfValueNull(plan.Iops) && !plan.Iops.Equal(state.Iops)) {
		numSpotVolume, err = core.UpdateVolumeAttributes(ctx, r.provider, deserializeUpdateNumspotVolume(plan), volumeID, stateVMID)
		if err != nil {
			response.Diagnostics.AddError("unable to update volume attributes", err.Error())
			return
		}
	}

	if !plan.LinkVm.VmId.Equal(state.LinkVm.VmId) || !plan.LinkVm.DeviceName.Equal(state.LinkVm.DeviceName) {
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

	newState := serializeNumSpotVolume(ctx, numSpotVolume, &response.Diagnostics, state.ReplaceVolumeOnDownsize)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete handles the deletion of a Volume.
func (r *volumeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_volume.VolumeModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVolume(ctx, r.provider, state.Id.ValueString(), state.LinkVm.VmId.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete volume", err.Error())
		return
	}
}

// serializeNumSpotVolume converts NumSpot API Volume data to Terraform Volume model.
// It handles the conversion of all Volume attributes including:
// - Basic Volume information
// - Linked volumes
// - VM attachments
// - Tags
func serializeNumSpotVolume(ctx context.Context, http *api.Volume, diags *diag.Diagnostics, ReplaceVolumeOnDownsize basetypes.BoolValue) resource_volume.VolumeModel {
	var (
		volumes = types.ListNull(resource_volume.LinkedVolumesValue{}.Type(ctx))
		tagsTf  types.Set
		linkVm  resource_volume.LinkVmValue
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
			linkVm, diagnostics = resource_volume.NewLinkVmValue(resource_volume.LinkVmValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"device_name": types.StringPointerValue((*http.LinkedVolumes)[0].DeviceName),
					"vm_id":       types.StringPointerValue((*http.LinkedVolumes)[0].VmId),
				})
			diags.Append(diagnostics...)
		}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_volume.VolumeModel{
		CreationDate:            types.StringValue(http.CreationDate.String()),
		Id:                      types.StringPointerValue(http.Id),
		Iops:                    utils.FromIntPtrToTfInt64(http.Iops),
		Size:                    utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:              types.StringPointerValue(http.SnapshotId),
		State:                   types.StringPointerValue(http.State),
		AvailabilityZoneName:    types.StringValue(utils.ConvertAzNamePtrToString(http.AvailabilityZoneName)),
		Type:                    types.StringPointerValue(http.Type),
		LinkedVolumes:           volumes,
		Tags:                    tagsTf,
		LinkVm:                  linkVm,
		ReplaceVolumeOnDownsize: ReplaceVolumeOnDownsize,
	}
}

// serializeLinkedVolumes converts NumSpot API LinkedVolume data to Terraform model.
func serializeLinkedVolumes(ctx context.Context, http api.LinkedVolume, diags *diag.Diagnostics) resource_volume.LinkedVolumesValue {
	value, diagnostics := resource_volume.NewLinkedVolumesValue(
		resource_volume.LinkedVolumesValue{}.AttributeTypes(ctx),
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

// deserializeCreateNumSpotVolume converts Terraform Volume model to NumSpot API request.
// It handles the conversion of Volume creation attributes including:
// - Size
// - Type
// - IOPS
// - Snapshot ID
func deserializeCreateNumSpotVolume(tf resource_volume.VolumeModel) api.CreateVolumeJSONRequestBody {
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

	return api.CreateVolumeJSONRequestBody{
		Iops:                 httpIOPS,
		Size:                 utils.FromTfInt64ToIntPtr(tf.Size),
		SnapshotId:           snapshotId,
		AvailabilityZoneName: api.AvailabilityZoneName(tf.AvailabilityZoneName.ValueString()),
		Type:                 tf.Type.ValueStringPointer(),
	}
}

// deserializeUpdateNumspotVolume converts Terraform Volume model to NumSpot API update request.
func deserializeUpdateNumspotVolume(tf resource_volume.VolumeModel) api.UpdateVolumeJSONRequestBody {
	var httpIOPS *int
	if !tf.Iops.IsUnknown() && !tf.Iops.IsNull() {
		httpIOPS = utils.FromTfInt64ToIntPtr(tf.Iops)
	}

	return api.UpdateVolumeJSONRequestBody{
		Iops:       httpIOPS,
		Size:       utils.FromTfInt64ToIntPtr(tf.Size),
		VolumeType: tf.Type.ValueStringPointer(),
	}
}

// volumeTags converts Terraform tags to NumSpot API format.
func volumeTags(ctx context.Context, tags types.Set) []api.ResourceTag {
	tfTags := make([]resource_volume.TagsValue, 0, len(tags.Elements()))
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
