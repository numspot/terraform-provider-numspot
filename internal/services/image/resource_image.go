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
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &ImageResource{}
	_ resource.ResourceWithConfigure   = &ImageResource{}
	_ resource.ResourceWithImportState = &ImageResource{}
)

type ImageResource struct {
	provider *client.NumSpotSDK
}

func NewImageResource() resource.Resource {
	return &ImageResource{}
}

func (r *ImageResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *ImageResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *ImageResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_image"
}

func (r *ImageResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = ImageResourceSchema(ctx)
}

func (r *ImageResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan ImageModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
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

func (r *ImageResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ImageModel

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

func (r *ImageResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan  ImageModel
		err          error
		numSpotImage *numspot.Image
	)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	imageID := state.Id.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

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

func (r *ImageResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteImage(ctx, r.provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete image", err.Error())
		return
	}
}

func deserializeAccess(accessValue AccessValue) *numspot.Access {
	if utils.IsTfValueNull(accessValue) {
		return nil
	}
	return &numspot.Access{
		IsPublic: accessValue.IsPublic.ValueBoolPointer(),
	}
}

func deserializeCreateNumSpotImage(tf ImageModel, diag *diag.Diagnostics) *numspot.CreateImageJSONRequestBody {
	blockDevicesMappingApi := make([]numspot.BlockDeviceMappingImage, 0, len(tf.BlockDeviceMappings.Elements()))
	for _, bdmTf := range tf.BlockDeviceMappings.Elements() {
		bdmTfRes, ok := bdmTf.(BlockDeviceMappingsValue)
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

	return &numspot.CreateImageJSONRequestBody{
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

func deserializeBlockDeviceMapping(bdm BlockDeviceMappingsValue) numspot.BlockDeviceMappingImage {
	attrtypes := bdm.Bsu.AttributeTypes(context.Background())
	attrVals := bdm.Bsu.Attributes()
	bsuTF, diags := NewBsuValue(attrtypes, attrVals)
	if diags.HasError() {
		return numspot.BlockDeviceMappingImage{}
	}
	bsu := deserializeBsuFromTf(bsuTF)

	return numspot.BlockDeviceMappingImage{
		Bsu:               bsu,
		DeviceName:        bdm.DeviceName.ValueStringPointer(),
		VirtualDeviceName: utils.FromTfStringToStringPtr(bdm.VirtualDeviceName),
	}
}

func deserializeBsuFromTf(bsu BsuValue) *numspot.BsuToCreate {
	if bsu.IsNull() || bsu.IsUnknown() {
		return nil
	}

	return &numspot.BsuToCreate{
		DeleteOnVmDeletion: bsu.DeleteOnVmDeletion.ValueBoolPointer(),
		Iops:               utils.FromTfInt64ToIntPtr(bsu.Iops),
		SnapshotId:         bsu.SnapshotId.ValueStringPointer(),
		VolumeSize:         utils.FromTfInt64ToIntPtr(bsu.VolumeSize),
		VolumeType:         bsu.VolumeType.ValueStringPointer(),
	}
}

func serializeNumSpotImage(ctx context.Context, plan ImageModel, image *numspot.Image, diags *diag.Diagnostics) *ImageModel {
	var (
		creationDateTf        types.String
		blockDeviceMappingsTf types.List
		productCodesTf        types.List
		stateCommentTf        StateCommentValue
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
		blockDeviceMappingsTf = types.ListNull(BlockDeviceMappingsValue{}.Type(ctx))
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
		stateCommentTf = NewStateCommentValueNull()
	}

	// Tags
	if image.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *image.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	access, diagnostics := NewAccessValue(AccessValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_public": types.BoolPointerValue(image.Access.IsPublic),
		},
	)

	diags.Append(diagnostics...)

	serializedImage := ImageModel{
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

func serializeStateComment(ctx context.Context, state numspot.StateComment, diags *diag.Diagnostics) StateCommentValue {
	stateCommentValue, diagnostics := NewStateCommentValue(
		StateCommentValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state_code":    types.StringPointerValue(state.StateCode),
			"state_message": types.StringPointerValue(state.StateMessage),
		},
	)
	diags.Append(diagnostics...)
	return stateCommentValue
}

func serializeBlockDeviceMapping(ctx context.Context, bdm numspot.BlockDeviceMappingImage, diags *diag.Diagnostics) BlockDeviceMappingsValue {
	bsu := serializeBsu(ctx, bdm.Bsu, diags)
	if diags.HasError() {
		return NewBlockDeviceMappingsValueNull()
	}

	bsuObjectValue, diagnostics := bsu.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return NewBlockDeviceMappingsValueNull()
	}

	blockDeviceMappingValue, diagnostics := NewBlockDeviceMappingsValue(
		BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":                 bsuObjectValue,
			"device_name":         types.StringPointerValue(bdm.DeviceName),
			"virtual_device_name": types.StringPointerValue(bdm.VirtualDeviceName),
		},
	)
	diags.Append(diagnostics...)
	return blockDeviceMappingValue
}

func serializeBsu(ctx context.Context, bsu *numspot.BsuToCreate, diags *diag.Diagnostics) BsuValue {
	if bsu == nil {
		return NewBsuValueNull()
	}

	bsuValue, diagnostics := NewBsuValue(
		BsuValue{}.AttributeTypes(ctx),
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
