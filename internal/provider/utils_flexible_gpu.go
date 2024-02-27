package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_flexible_gpu"
)

func FlexibleGpuFromHttpToTf(http *api.FlexibleGpu) resource_flexible_gpu.FlexibleGpuModel {
	return resource_flexible_gpu.FlexibleGpuModel{
		DeleteOnVmDeletion: types.BoolPointerValue(http.DeleteOnVmDeletion),
		Generation:         types.StringPointerValue(http.Generation),
		Id:                 types.StringPointerValue(http.Id),
		ModelName:          types.StringPointerValue(http.ModelName),
		State:              types.StringPointerValue(http.State),
		SubregionName:      types.StringPointerValue(http.AvailabilityZoneName),
		VmId:               types.StringPointerValue(http.VmId),
	}
}

func FlexibleGpuFromTfToCreateRequest(tf *resource_flexible_gpu.FlexibleGpuModel) api.CreateFlexibleGpuJSONRequestBody {
	return api.CreateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion:   tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generation:           tf.Generation.ValueStringPointer(),
		ModelName:            tf.ModelName.ValueString(),
		AvailabilityZoneName: tf.SubregionName.ValueString(),
	}
}
