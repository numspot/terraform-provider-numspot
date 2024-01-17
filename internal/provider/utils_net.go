package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_net"
)

func NetFromTfToHttp(tf resource_net.NetModel) *api.NetSchema {
	return &api.NetSchema{}
}

func NetFromHttpToTf(http *api.NetSchema) resource_net.NetModel {
	return resource_net.NetModel{}
}

func NetFromTfToCreateRequest(tf resource_net.NetModel) api.CreateNetJSONRequestBody {
	return api.CreateNetJSONRequestBody{}
}
