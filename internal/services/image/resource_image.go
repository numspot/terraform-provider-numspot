package image

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/image/resource_image"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewImageResource() resource.Resource {
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
	response.TypeName = request.ProviderTypeName + "_image"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_image.ImageResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_image.ImageModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := imageTags(ctx, plan.Tags)
	body := deserializeCreateNumSpotImage(plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	numSpotImage, err := core.CreateImage(ctx, r.provider, *body, tagsValue, deserializeAccess(plan.Access))
	if err != nil {
		response.Diagnostics.AddError("unable to create image", err.Error())
		return
	}

	state := serializeNumSpotImage(ctx, plan, numSpotImage, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_image.ImageModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	imageID := state.Id.ValueString()

	numSpotImage, err := core.ReadImageWithID(ctx, r.provider, imageID)
	if err != nil {
		response.Diagnostics.AddError("unable to read image", err.Error())
		return
	}

	newState := serializeNumSpotImage(ctx, state, numSpotImage, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan  resource_image.ImageModel
		err          error
		numSpotImage *api.Image
	)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	imageID := state.Id.ValueString()
	planTags := imageTags(ctx, plan.Tags)
	stateTags := imageTags(ctx, state.Tags)

	if !state.Tags.Equal(plan.Tags) {
		numSpotImage, err = core.UpdateImageTags(ctx, r.provider, imageID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update image tags", err.Error())
			return
		}
	}

	if !state.Access.Equal(plan.Access) {
		numSpotImage, err = core.UpdateImageAccess(ctx, r.provider, imageID, *deserializeAccess(plan.Access))
		if err != nil {
			response.Diagnostics.AddError("unable to update image access", err.Error())
			return
		}
	}

	newState := serializeNumSpotImage(ctx, state, numSpotImage, &response.Diagnostics)
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_image.ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteImage(ctx, r.provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete image", err.Error())
		return
	}
}

func deserializeAccess(accessValue resource_image.AccessValue) *api.Access {
	if utils.IsTfValueNull(accessValue) {
		return nil
	}
	return &api.Access{
		IsPublic: accessValue.IsPublic.ValueBoolPointer(),
	}
}

func deserializeCreateNumSpotImage(tf resource_image.ImageModel, diag *diag.Diagnostics) *api.CreateImageJSONRequestBody {
	blockDevicesMappingApi := make([]api.BlockDeviceMappingImage, 0, len(tf.BlockDeviceMappings.Elements()))
	for _, bdmTf := range tf.BlockDeviceMappings.Elements() {
		bdmTfRes, ok := bdmTf.(resource_image.BlockDeviceMappingsValue)
		if !ok {
			diag.AddError("unable to cast block device mapping resource", "")
			return nil
		}

		bdmApi := deserializeBlockDeviceMapping(bdmTfRes)
		blockDevicesMappingApi = append(blockDevicesMappingApi, bdmApi)
	}

	productCodesApi := make([]string, 0, len(tf.ProductCodes.Elements()))
	for _, pcTf := range tf.ProductCodes.Elements() {
		pcTfStr, ok := pcTf.(types.String)
		if !ok {
			diag.AddError("unable to cast product code to string", "")
			return nil
		}
		if pcTfStr.IsUnknown() || pcTfStr.IsNull() {
			continue
		}
		productCodesApi = append(productCodesApi, pcTfStr.ValueString())
	}

	return &api.CreateImageJSONRequestBody{
		Architecture:        utils.FromTfStringToStringPtr(tf.Architecture),
		BlockDeviceMappings: &blockDevicesMappingApi,
		Description:         utils.FromTfStringToStringPtr(tf.Description),
		Name:                utils.FromTfStringToStringPtr(tf.Name),
		NoReboot:            utils.FromTfBoolToBoolPtr(tf.NoReboot),
		ProductCodes:        &productCodesApi,
		RootDeviceName:      utils.FromTfStringToStringPtr(tf.RootDeviceName),
		SourceImageId:       utils.FromTfStringToStringPtr(tf.SourceImageId),
		SourceRegionName:    utils.FromTfStringToStringPtr(tf.SourceRegionName),
		VmId:                utils.FromTfStringToStringPtr(tf.VmId),
	}
}

func deserializeBlockDeviceMapping(bdm resource_image.BlockDeviceMappingsValue) api.BlockDeviceMappingImage {
	attrtypes := bdm.Bsu.AttributeTypes(context.Background())
	attrVals := bdm.Bsu.Attributes()
	bsuTF, diags := resource_image.NewBsuValue(attrtypes, attrVals)
	if diags.HasError() {
		return api.BlockDeviceMappingImage{}
	}
	bsu := deserializeBsuFromTf(bsuTF)

	return api.BlockDeviceMappingImage{
		Bsu:               bsu,
		DeviceName:        bdm.DeviceName.ValueStringPointer(),
		VirtualDeviceName: utils.FromTfStringToStringPtr(bdm.VirtualDeviceName),
	}
}

func deserializeBsuFromTf(bsu resource_image.BsuValue) *api.BsuToCreate {
	if bsu.IsNull() || bsu.IsUnknown() {
		return nil
	}

	return &api.BsuToCreate{
		DeleteOnVmDeletion: bsu.DeleteOnVmDeletion.ValueBoolPointer(),
		Iops:               utils.FromTfInt64ToIntPtr(bsu.Iops),
		SnapshotId:         bsu.SnapshotId.ValueStringPointer(),
		VolumeSize:         utils.FromTfInt64ToIntPtr(bsu.VolumeSize),
		VolumeType:         bsu.VolumeType.ValueStringPointer(),
	}
}

func serializeNumSpotImage(ctx context.Context, plan resource_image.ImageModel, image *api.Image, diags *diag.Diagnostics) *resource_image.ImageModel {
	var (
		creationDateTf        types.String
		blockDeviceMappingsTf types.List
		productCodesTf        types.List
		stateCommentTf        resource_image.StateCommentValue
		tagsTf                types.List
	)

	// Creation Date
	if image.CreationDate != nil {
		date := *image.CreationDate
		creationDateTf = types.StringValue(date.Format(time.RFC3339))
	}

	// Block Device Mapping
	if image.BlockDeviceMappings != nil {
		blockDeviceMappingsTf = utils.GenericListToTfListValue(
			ctx,
			serializeBlockDeviceMapping,
			*image.BlockDeviceMappings,
			diags,
		)
		if diags.HasError() {
			return nil
		}
	} else {
		blockDeviceMappingsTf = types.ListNull(resource_image.BlockDeviceMappingsValue{}.Type(ctx))
	}

	// Product Codes
	if image.ProductCodes != nil {
		productCodesTf = utils.StringListToTfListValue(ctx, *image.ProductCodes, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		productCodesTf = types.ListNull(types.StringType)
	}

	// State Comment
	if image.StateComment != nil {
		stateCommentTf = serializeStateComment(ctx, *image.StateComment, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		stateCommentTf = resource_image.NewStateCommentValueNull()
	}

	// Tags
	if image.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *image.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	access, diagnostics := resource_image.NewAccessValue(resource_image.AccessValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_public": types.BoolPointerValue(image.Access.IsPublic),
		},
	)

	diags.Append(diagnostics...)

	serializedImage := resource_image.ImageModel{
		Architecture:        types.StringPointerValue(image.Architecture),
		CreationDate:        creationDateTf,
		Description:         types.StringPointerValue(image.Description),
		Id:                  types.StringPointerValue(image.Id),
		Name:                types.StringPointerValue(image.Name),
		RootDeviceName:      types.StringPointerValue(image.RootDeviceName),
		RootDeviceType:      types.StringPointerValue(image.RootDeviceType),
		State:               types.StringPointerValue(image.State),
		Type:                types.StringPointerValue(image.Type),
		StateComment:        stateCommentTf,
		ProductCodes:        productCodesTf,
		BlockDeviceMappings: blockDeviceMappingsTf,
		Tags:                tagsTf,
		Access:              access,
	}

	serializedImage.SourceImageId = utils.FromTfStringValueToTfOrNull(plan.SourceImageId)
	serializedImage.SourceRegionName = utils.FromTfStringValueToTfOrNull(plan.SourceRegionName)
	serializedImage.VmId = utils.FromTfStringValueToTfOrNull(plan.VmId)
	serializedImage.NoReboot = utils.FromTfBoolValueToTfOrNull(plan.NoReboot)

	return &serializedImage
}

func serializeStateComment(ctx context.Context, state api.StateComment, diags *diag.Diagnostics) resource_image.StateCommentValue {
	stateCommentValue, diagnostics := resource_image.NewStateCommentValue(
		resource_image.StateCommentValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state_code":    types.StringPointerValue(state.StateCode),
			"state_message": types.StringPointerValue(state.StateMessage),
		},
	)
	diags.Append(diagnostics...)
	return stateCommentValue
}

func serializeBlockDeviceMapping(ctx context.Context, bdm api.BlockDeviceMappingImage, diags *diag.Diagnostics) resource_image.BlockDeviceMappingsValue {
	bsu := serializeBsu(ctx, bdm.Bsu, diags)
	if diags.HasError() {
		return resource_image.NewBlockDeviceMappingsValueNull()
	}

	bsuObjectValue, diagnostics := bsu.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return resource_image.NewBlockDeviceMappingsValueNull()
	}

	blockDeviceMappingValue, diagnostics := resource_image.NewBlockDeviceMappingsValue(
		resource_image.BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":                 bsuObjectValue,
			"device_name":         types.StringPointerValue(bdm.DeviceName),
			"virtual_device_name": types.StringPointerValue(bdm.VirtualDeviceName),
		},
	)
	diags.Append(diagnostics...)
	return blockDeviceMappingValue
}

func serializeBsu(ctx context.Context, bsu *api.BsuToCreate, diags *diag.Diagnostics) resource_image.BsuValue {
	if bsu == nil {
		return resource_image.NewBsuValueNull()
	}

	bsuValue, diagnostics := resource_image.NewBsuValue(
		resource_image.BsuValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(bsu.DeleteOnVmDeletion),
			"iops":                  utils.FromIntPtrToTfInt64(bsu.Iops),
			"snapshot_id":           types.StringPointerValue(bsu.SnapshotId),
			"volume_size":           utils.FromIntPtrToTfInt64(bsu.VolumeSize),
			"volume_type":           types.StringPointerValue(bsu.VolumeType),
		},
	)
	diags.Append(diagnostics...)
	return bsuValue
}

func imageTags(ctx context.Context, tags types.List) []api.ResourceTag {
	tfTags := make([]resource_image.TagsValue, 0, len(tags.Elements()))
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
