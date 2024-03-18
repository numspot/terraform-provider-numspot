package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_image"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func bsuFromTf(bsu resource_image.BsuValue) *api.BsuToCreate {
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

func blockDeviceMappingFromTf(ctx context.Context, bdm resource_image.BlockDeviceMappingsValue) api.BlockDeviceMappingImage {
	bsuTf := resource_image.BsuValue{}
	bsu := bsuFromTf(bsuTf)

	return api.BlockDeviceMappingImage{
		Bsu:               bsu,
		DeviceName:        bdm.DeviceName.ValueStringPointer(),
		VirtualDeviceName: bdm.VirtualDeviceName.ValueStringPointer(),
	}
}

func bsuFromApi(ctx context.Context, bsu *api.BsuToCreate) (resource_image.BsuValue, diag.Diagnostics) {
	if bsu == nil {
		return resource_image.NewBsuValueNull(), nil
	}

	return resource_image.NewBsuValue(
		resource_image.BsuValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(bsu.DeleteOnVmDeletion),
			"iops":                  utils.FromIntPtrToTfInt64(bsu.Iops),
			"snapshot_id":           types.StringPointerValue(bsu.SnapshotId),
			"volume_size":           utils.FromIntPtrToTfInt64(bsu.VolumeSize),
			"volume_type":           types.StringPointerValue(bsu.VolumeType),
		},
	)
}

func blockDeviceMappingFromApi(ctx context.Context, bdm api.BlockDeviceMappingImage) (resource_image.BlockDeviceMappingsValue, diag.Diagnostics) {
	bsu, diagnostics := bsuFromApi(ctx, bdm.Bsu)
	if diagnostics.HasError() {
		return resource_image.NewBlockDeviceMappingsValueNull(), diagnostics
	}

	bsuObjectValue, diagnostics := bsu.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return resource_image.NewBlockDeviceMappingsValueNull(), diagnostics
	}

	return resource_image.NewBlockDeviceMappingsValue(
		resource_image.BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":                 bsuObjectValue,
			"device_name":         types.StringPointerValue(bdm.DeviceName),
			"virtual_device_name": types.StringPointerValue(bdm.VirtualDeviceName),
		},
	)
}

func stateCommentFromApi(ctx context.Context, state api.StateComment) (resource_image.StateCommentValue, diag.Diagnostics) {
	return resource_image.NewStateCommentValue(
		resource_image.StateCommentValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state_code":    types.StringPointerValue(state.StateCode),
			"state_message": types.StringPointerValue(state.StateMessage),
		},
	)
}

func ImageFromHttpToTf(ctx context.Context, http *api.Image) (*resource_image.ImageModel, diag.Diagnostics) {
	var (
		creationDateTf        types.String
		blockDeviceMappingsTf types.List
		productCodesTf        types.List
		diagnostics           diag.Diagnostics
		stateCommentTf        resource_image.StateCommentValue
	)

	// Creation Date
	if http.CreationDate != nil {
		date := *http.CreationDate
		creationDateTf = types.StringValue(date.Format(time.RFC3339))
	}

	// Block Device Mapping
	if http.BlockDeviceMappings != nil {
		blockDeviceMappingsTf, diagnostics = utils.GenericListToTfListValue(
			ctx,
			resource_image.BlockDeviceMappingsValue{},
			blockDeviceMappingFromApi,
			*http.BlockDeviceMappings,
		)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
	} else {
		blockDeviceMappingsTf = types.ListNull(resource_image.BlockDeviceMappingsValue{}.Type(ctx))
	}

	// Product Codes
	if http.ProductCodes != nil {
		productCodesTf, diagnostics = utils.StringListToTfListValue(ctx, *http.ProductCodes)
	} else {
		productCodesTf = types.ListNull(types.StringType)
	}

	// State Comment
	if http.StateComment != nil {
		stateCommentTf, diagnostics = stateCommentFromApi(ctx, *http.StateComment)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
	} else {
		stateCommentTf = resource_image.NewStateCommentValueNull()
	}

	return &resource_image.ImageModel{
		Architecture:   types.StringPointerValue(http.Architecture),
		CreationDate:   creationDateTf,
		Description:    types.StringPointerValue(http.Description),
		Id:             types.StringPointerValue(http.Id),
		Name:           types.StringPointerValue(http.Name),
		RootDeviceName: types.StringPointerValue(http.RootDeviceName),
		RootDeviceType: types.StringPointerValue(http.RootDeviceType),
		State:          types.StringPointerValue(http.State),
		Type:           types.StringPointerValue(http.Type),

		// TODO: Handle those fields after this function call
		// SourceImageId:    types.StringPointerValue(http.S),
		// SourceRegionName: types.StringPointerValue(http.),
		// VmId:  types.StringPointerValue(http.),
		// NoReboot:       types.BoolPointerValue(http.),

		//
		StateComment: stateCommentTf,
		//
		ProductCodes:        productCodesTf,
		BlockDeviceMappings: blockDeviceMappingsTf,
	}, diagnostics
}

func ImageFromTfToCreateRequest(ctx context.Context, tf *resource_image.ImageModel, diag *diag.Diagnostics) *api.CreateImageJSONRequestBody {
	blockDevicesMappingApi := make([]api.BlockDeviceMappingImage, 0, len(tf.BlockDeviceMappings.Elements()))
	for _, bdmTf := range tf.BlockDeviceMappings.Elements() {
		bdmTfRes, ok := bdmTf.(resource_image.BlockDeviceMappingsValue)
		if !ok {
			diag.AddError("Failed to cast block device mapping resource", "")
			return nil
		}

		bdmApi := blockDeviceMappingFromTf(ctx, bdmTfRes)
		blockDevicesMappingApi = append(blockDevicesMappingApi, bdmApi)
	}

	productCodesApi := make([]string, 0, len(tf.ProductCodes.Elements()))
	for _, pcTf := range tf.ProductCodes.Elements() {
		pcTfStr, ok := pcTf.(types.String)
		if !ok {
			diag.AddError("Failed to cast product code to string", "")
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
