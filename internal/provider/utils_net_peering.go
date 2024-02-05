package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_net_peering"
)

func NetPeeringFromTfToHttp(tf *resource_net_peering.NetPeeringModel) *api.NetPeeringSchema {
	return &api.NetPeeringSchema{}
}

func NetPeeringFromHttpToTf(http *api.NetPeeringSchema) resource_net_peering.NetPeeringModel {
	return resource_net_peering.NetPeeringModel{}
}

func NetPeeringFromTfToCreateRequest(tf *resource_net_peering.NetPeeringModel) api.CreateNetPeeringJSONRequestBody {
	return api.CreateNetPeeringJSONRequestBody{}
}
