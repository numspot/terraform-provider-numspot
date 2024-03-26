package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_flexible_gpu"
)

func FlexibleGpuFromHttpToTf(http *iaas.FlexibleGpu) resource_flexible_gpu.FlexibleGpuModel {
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

func FlexibleGpuFromTfToCreateRequest(tf *resource_flexible_gpu.FlexibleGpuModel) iaas.CreateFlexibleGpuJSONRequestBody {
	return iaas.CreateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion:   tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generation:           tf.Generation.ValueStringPointer(),
		ModelName:            tf.ModelName.ValueString(),
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueString(),
	}
}
