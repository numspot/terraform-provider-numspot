package volume

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/tags"
	"github.com/numspot/terraform-provider-numspot/internal/services/volume/resource_volume"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &volumeResource{}
	_ resource.ResourceWithConfigure   = &volumeResource{}
	_ resource.ResourceWithImportState = &volumeResource{}
)

// volumeResource is the Terraform resource implementation for numspot_volume.
type volumeResource struct {
	services.BaseResource
}

// NewVolumeResource returns a new volumeResource instance.
func NewVolumeResource() resource.Resource {
	return &volumeResource{}
}

// Metadata sets the resource type name.
func (r *volumeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_volume"
}

// Schema defines the schema for the resource.
func (r *volumeResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_volume.VolumeResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *volumeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_volume.VolumeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := volumeTags(ctx, plan.Tags)
	numspotVolume, err := core.CreateVolume(
		ctx,
		r.Provider,
		deserializeCreateVolumeRequest(&plan),
		plan.LinkVm.VmId.ValueString(),
		plan.LinkVm.DeviceName.ValueString(),
		tagsValue,
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create volume", err.Error())
		return
	}

	state := serializeVolume(ctx, numspotVolume, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	state.ReplaceVolumeOnDownsize = utils.PreserveFromState(plan.ReplaceVolumeOnDownsize)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *volumeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_volume.VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotVolume, err := core.ReadVolume(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read volume", err.Error())
		return
	}

	newState := serializeVolume(ctx, numspotVolume, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.ReplaceVolumeOnDownsize = state.ReplaceVolumeOnDownsize

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *volumeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err           error
		state, plan   resource_volume.VolumeModel
		numspotVolume *api.Volume
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	volumeID := state.Id.ValueString()
	planTags := volumeTags(ctx, plan.Tags)
	stateTags := volumeTags(ctx, state.Tags)

	if !plan.Size.Equal(state.Size) || !plan.Type.Equal(state.Type) || !plan.Iops.Equal(state.Iops) {
		numspotVolume, err = core.UpdateVolumeAttributes(
			ctx,
			r.Provider,
			deserializeUpdateVolume(&plan),
			state.Id.ValueString(),
			state.LinkVm.VmId.ValueString(),
			state.Size.ValueInt64(),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update volume", err.Error())
			return
		}
	}

	if !plan.LinkVm.Equal(state.LinkVm) {
		numspotVolume, err = core.UpdateVolumeLink(
			ctx,
			r.Provider,
			state.Id.ValueString(),
			state.LinkVm.VmId.ValueString(),
			plan.LinkVm.VmId.ValueString(),
			plan.LinkVm.DeviceName.ValueString(),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update volume", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numspotVolume, err = core.UpdateVolumeTags(ctx, r.Provider, volumeID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update volume tags", err.Error())
			return
		}
	}

	newState := serializeVolume(ctx, numspotVolume, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.ReplaceVolumeOnDownsize = state.ReplaceVolumeOnDownsize

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *volumeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_volume.VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVolume(ctx, r.Provider, state.Id.ValueString(), state.LinkVm.VmId.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete volume", err.Error())
		return
	}
}

func serializeVolume(ctx context.Context, http *api.Volume, diags *diag.Diagnostics) resource_volume.VolumeModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil Volume response", "the API returned an empty Volume")
		return resource_volume.VolumeModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_volume.VolumeModel{
		AvailabilityZoneName: types.StringValue(utils.ConvertAzNamePtrToString(http.AvailabilityZoneName)),
		CreationDate:         utils.TimePtrToTfString(http.CreationDate),
		Id:                   types.StringPointerValue(http.Id),
		Iops:                 utils.FromIntPtrToTfInt64(http.Iops),
		LinkVm:               linkVmFromVolume(ctx, http.LinkedVolumes, diags),
		LinkedVolumes:        utils.GenericListToTfListValue(ctx, serializeLinkedVolume, *http.LinkedVolumes, diags),
		Size:                 utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:           types.StringPointerValue(http.SnapshotId),
		State:                types.StringPointerValue(http.State),
		Type:                 types.StringPointerValue(http.Type),
		Tags:                 tagsTf,
	}
}

func deserializeCreateVolumeRequest(tf *resource_volume.VolumeModel) *api.CreateVolumeJSONRequestBody {
	return &api.CreateVolumeJSONRequestBody{
		AvailabilityZoneName: api.AvailabilityZoneName(tf.AvailabilityZoneName.ValueString()),
		Iops:                 utils.FromTfInt64ToIntPtr(tf.Iops),
		Size:                 utils.FromTfInt64ToIntPtr(tf.Size),
		SnapshotId:           utils.FromTfStringToStringPtr(tf.SnapshotId),
		Type:                 utils.FromTfStringToStringPtr(tf.Type),
	}
}

func volumeTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_volume.TagsValue, 0, len(tagsSet.Elements()))
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

func serializeLinkedVolume(ctx context.Context, item *api.LinkedVolume, diags *diag.Diagnostics) resource_volume.LinkedVolumesValue {
	value, d := resource_volume.NewLinkedVolumesValue(resource_volume.LinkedVolumesValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(item.DeleteOnVmDeletion),
		"device_name":           types.StringPointerValue(item.DeviceName),
		"id":                    types.StringPointerValue(item.Id),
		"state":                 types.StringPointerValue(item.State),
		"vm_id":                 types.StringPointerValue(item.VmId),
	})
	diags.Append(d...)
	return value
}
