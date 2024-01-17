package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_net_access_point"
)

func NetAccessPointFromTfToHttp(tf resource_net_access_point.NetAccessPointModel) *api.NetAccessPointSchema {
	return &api.NetAccessPointSchema{}
}

func NetAccessPointFromHttpToTf(http *api.NetAccessPointSchema) resource_net_access_point.NetAccessPointModel {
	return resource_net_access_point.NetAccessPointModel{}
}

func NetAccessPointFromTfToCreateRequest(tf resource_net_access_point.NetAccessPointModel) api.CreateNetAccessPointJSONRequestBody {
	return api.CreateNetAccessPointJSONRequestBody{}
}
