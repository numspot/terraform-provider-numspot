package image

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/image/resource_image"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &imageResource{}
	_ resource.ResourceWithConfigure   = &imageResource{}
	_ resource.ResourceWithImportState = &imageResource{}
)

// imageResource is the Terraform resource implementation for numspot_image.
type imageResource struct {
	services.BaseResource
}

// NewImageResource returns a new imageResource instance.
func NewImageResource() resource.Resource {
	return &imageResource{}
}

// Metadata sets the resource type name.
func (r *imageResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_image"
}

// Schema defines the schema for the resource.
func (r *imageResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_image.ImageResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *imageResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_image.ImageModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := imageTags(ctx, plan.Tags)
	numSpotImage, err := core.CreateImage(ctx, r.Provider, deserializeCreateImageRequest(ctx, &plan, &response.Diagnostics), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create image", err.Error())
		return
	}

	state := serializeImage(ctx, numSpotImage, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	state.SourceImageId = utils.PreserveFromState(plan.SourceImageId)
	state.SourceRegionName = utils.PreserveFromState(plan.SourceRegionName)
	state.VmId = utils.PreserveFromState(plan.VmId)
	state.NoReboot = utils.PreserveFromState(plan.NoReboot)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *imageResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_image.ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotImage, err := core.ReadImage(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read image", err.Error())
		return
	}

	newState := serializeImage(ctx, numSpotImage, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.SourceImageId = state.SourceImageId
	newState.SourceRegionName = state.SourceRegionName
	newState.VmId = state.VmId
	newState.NoReboot = state.NoReboot

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *imageResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err          error
		state, plan  resource_image.ImageModel
		numSpotImage *api.Image
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	imageID := state.Id.ValueString()
	planTags := imageTags(ctx, plan.Tags)
	stateTags := imageTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotImage, err = core.UpdateImageTags(ctx, r.Provider, imageID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update image tags", err.Error())
			return
		}
	}

	newState := serializeImage(ctx, numSpotImage, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.SourceImageId = state.SourceImageId
	newState.SourceRegionName = state.SourceRegionName
	newState.VmId = state.VmId
	newState.NoReboot = state.NoReboot

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *imageResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_image.ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteImage(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete image", err.Error())
		return
	}
}

func serializeImage(ctx context.Context, http *api.Image, diags *diag.Diagnostics) resource_image.ImageModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil Image response", "the API returned an empty Image")
		return resource_image.ImageModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_image.ImageModel{
		Access:              serializeAccess(ctx, http.Access, diags),
		Architecture:        types.StringPointerValue(http.Architecture),
		BlockDeviceMappings: utils.GenericListToTfListValue(ctx, serializeBlockDeviceMapping, *http.BlockDeviceMappings, diags),
		CreationDate:        utils.TimePtrToTfString(http.CreationDate),
		Description:         types.StringPointerValue(http.Description),
		Id:                  types.StringPointerValue(http.Id),
		Name:                types.StringPointerValue(http.Name),
		ProductCodes:        utils.FromStringListPointerToTfStringList(ctx, http.ProductCodes, diags),
		RootDeviceName:      types.StringPointerValue(http.RootDeviceName),
		RootDeviceType:      types.StringPointerValue(http.RootDeviceType),
		State:               types.StringPointerValue(http.State),
		StateComment:        serializeStateComment(ctx, http.StateComment, diags),
		Type:                types.StringPointerValue(http.Type),
		Tags:                tagsTf,
	}
}

func deserializeCreateImageRequest(ctx context.Context, tf *resource_image.ImageModel, diags *diag.Diagnostics) *api.CreateImageJSONRequestBody {
	return &api.CreateImageJSONRequestBody{
		Architecture:        utils.FromTfStringToStringPtr(tf.Architecture),
		BlockDeviceMappings: utils.PointerOf(deserializeBlockDeviceMappings(ctx, tf.BlockDeviceMappings)),
		Description:         utils.FromTfStringToStringPtr(tf.Description),
		Name:                utils.FromTfStringToStringPtr(tf.Name),
		NoReboot:            utils.FromTfBoolToBoolPtr(tf.NoReboot),
		ProductCodes:        utils.TfStringListToStringPtrList(ctx, tf.ProductCodes, diags),
		RootDeviceName:      utils.FromTfStringToStringPtr(tf.RootDeviceName),
		SourceImageId:       utils.FromTfStringToStringPtr(tf.SourceImageId),
		SourceRegionName:    utils.FromTfStringToStringPtr(tf.SourceRegionName),
		VmId:                utils.FromTfStringToStringPtr(tf.VmId),
	}
}

func imageTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_image.TagsValue, 0, len(tagsSet.Elements()))
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

func serializeAccess(ctx context.Context, item *api.Access, diags *diag.Diagnostics) resource_image.AccessValue {
	if item == nil {
		return resource_image.NewAccessValueNull()
	}
	value, d := resource_image.NewAccessValue(resource_image.AccessValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"is_public": types.BoolPointerValue(item.IsPublic),
	})
	diags.Append(d...)
	return value
}

func serializeBsu(ctx context.Context, item *api.BsuToCreate, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(resource_image.BsuValue{}.AttributeTypes(ctx))
	}
	value, d := resource_image.NewBsuValue(resource_image.BsuValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(item.DeleteOnVmDeletion),
		"iops":                  utils.FromIntPtrToTfInt64(item.Iops),
		"snapshot_id":           types.StringPointerValue(item.SnapshotId),
		"volume_size":           utils.FromIntPtrToTfInt64(item.VolumeSize),
		"volume_type":           types.StringPointerValue(item.VolumeType),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func deserializeBsu(obj types.Object) *api.BsuToCreate {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	// obj.As cannot target the generated attr.Value structs (the framework
	// reflect layer rejects custom attr.Value targets), so map the attribute
	// values manually.
	attrs := obj.Attributes()
	getString := func(name string) *string {
		v, ok := attrs[name].(types.String)
		if !ok || v.IsNull() || v.IsUnknown() {
			return nil
		}
		return v.ValueStringPointer()
	}
	getInt := func(name string) *int {
		v, ok := attrs[name].(types.Int64)
		if !ok || v.IsNull() || v.IsUnknown() {
			return nil
		}
		return utils.FromTfInt64ToIntPtr(v)
	}
	getBool := func(name string) *bool {
		v, ok := attrs[name].(types.Bool)
		if !ok || v.IsNull() || v.IsUnknown() {
			return nil
		}
		return v.ValueBoolPointer()
	}
	return &api.BsuToCreate{
		DeleteOnVmDeletion: getBool("delete_on_vm_deletion"),
		Iops:               getInt("iops"),
		SnapshotId:         getString("snapshot_id"),
		VolumeSize:         getInt("volume_size"),
		VolumeType:         getString("volume_type"),
	}
}

func serializeBlockDeviceMapping(
	ctx context.Context,
	item *api.BlockDeviceMappingImage,
	diags *diag.Diagnostics,
) resource_image.BlockDeviceMappingsValue {
	value, d := resource_image.NewBlockDeviceMappingsValue(resource_image.BlockDeviceMappingsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"bsu":                 serializeBsu(ctx, item.Bsu, diags),
		"device_name":         types.StringPointerValue(item.DeviceName),
		"virtual_device_name": types.StringPointerValue(item.VirtualDeviceName),
	})
	diags.Append(d...)
	return value
}

func deserializeBlockDeviceMappings(ctx context.Context, tfBlockDeviceMappings types.List) []api.BlockDeviceMappingImage {
	tfItems := make([]resource_image.BlockDeviceMappingsValue, 0, len(tfBlockDeviceMappings.Elements()))
	tfBlockDeviceMappings.ElementsAs(ctx, &tfItems, false)

	items := make([]api.BlockDeviceMappingImage, 0, len(tfItems))
	for i := range tfItems {
		e := &tfItems[i]
		items = append(items, api.BlockDeviceMappingImage{
			Bsu:               deserializeBsu(e.Bsu),
			DeviceName:        e.DeviceName.ValueStringPointer(),
			VirtualDeviceName: utils.FromTfStringToStringPtr(e.VirtualDeviceName),
		})
	}
	return items
}

func serializeStateComment(ctx context.Context, item *api.StateComment, diags *diag.Diagnostics) resource_image.StateCommentValue {
	if item == nil {
		return resource_image.NewStateCommentValueNull()
	}
	value, d := resource_image.NewStateCommentValue(resource_image.StateCommentValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"state_code":    types.StringPointerValue(item.StateCode),
		"state_message": types.StringPointerValue(item.StateMessage),
	})
	diags.Append(d...)
	return value
}
