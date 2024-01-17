package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_flexible_gpu"
)

func FlexibleGpuFromTfToHttp(tf resource_flexible_gpu.FlexibleGpuModel) *api.FlexibleGpuSchema {
	return &api.FlexibleGpuSchema{}
}

func FlexibleGpuFromHttpToTf(http *api.FlexibleGpuSchema) resource_flexible_gpu.FlexibleGpuModel {
	return resource_flexible_gpu.FlexibleGpuModel{}
}

func FlexibleGpuFromTfToCreateRequest(tf resource_flexible_gpu.FlexibleGpuModel) api.CreateFlexibleGpuJSONRequestBody {
	return api.CreateFlexibleGpuJSONRequestBody{}
}
