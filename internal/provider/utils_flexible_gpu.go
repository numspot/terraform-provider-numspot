package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_flexible_gpu"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_flexible_gpu"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func FlexibleGpuFromHttpToTf(http *numspot.FlexibleGpu) resource_flexible_gpu.FlexibleGpuModel {
	return resource_flexible_gpu.FlexibleGpuModel{
		DeleteOnVmDeletion:   types.BoolPointerValue(http.DeleteOnVmDeletion),
		Generation:           types.StringPointerValue(http.Generation),
		Id:                   types.StringPointerValue(http.Id),
		ModelName:            types.StringPointerValue(http.ModelName),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		VmId:                 types.StringPointerValue(http.VmId),
	}
}

func FlexibleGpuFromTfToCreateRequest(tf *resource_flexible_gpu.FlexibleGpuModel) numspot.CreateFlexibleGpuJSONRequestBody {
	return numspot.CreateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion:   tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generation:           tf.Generation.ValueStringPointer(),
		ModelName:            tf.ModelName.ValueString(),
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueString(),
	}
}

func FlexibleGpuFromTfToUpdateRequest(tf *resource_flexible_gpu.FlexibleGpuModel) numspot.UpdateFlexibleGpuJSONRequestBody {
	return numspot.UpdateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.DeleteOnVmDeletion),
	}
}

func FlexibleGpusFromTfToAPIReadParams(ctx context.Context, tf FlexibleGpuModel) numspot.ReadFlexibleGpusParams {
	return numspot.ReadFlexibleGpusParams{
		States:                utils.TfStringListToStringPtrList(ctx, tf.States),
		Ids:                   utils.TfStringListToStringPtrList(ctx, tf.Ids),
		AvailabilityZoneNames: utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
		DeleteOnVmDeletion:    utils.FromTfBoolToBoolPtr(tf.DeleteOnVmDeletion),
		Generations:           utils.TfStringListToStringPtrList(ctx, tf.Generations),
		ModelNames:            utils.TfStringListToStringPtrList(ctx, tf.ModelNames),
		VmIds:                 utils.TfStringListToStringPtrList(ctx, tf.VmIds),
	}
}

func FlexibleGpusFromHttpToTfDatasource(ctx context.Context, http *numspot.FlexibleGpu) (*datasource_flexible_gpu.FlexibleGpuModel, diag.Diagnostics) {
	return &datasource_flexible_gpu.FlexibleGpuModel{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Id:                   types.StringPointerValue(http.Id),
		State:                types.StringPointerValue(http.State),
		DeleteOnVmDeletion:   types.BoolPointerValue(http.DeleteOnVmDeletion),
		Generation:           types.StringPointerValue(http.Generation),
		ModelName:            types.StringPointerValue(http.ModelName),
		VmId:                 types.StringPointerValue(http.VmId),
	}, nil
}

func LinkFlexibleGpuFromTfToCreateRequest(tf *resource_flexible_gpu.FlexibleGpuModel) numspot.LinkFlexibleGpuJSONRequestBody {
	vmId := utils.FromTfStringToStringPtr(tf.VmId)
	return numspot.LinkFlexibleGpuJSONRequestBody{
		VmId: utils.GetPtrValue(vmId),
	}
}
