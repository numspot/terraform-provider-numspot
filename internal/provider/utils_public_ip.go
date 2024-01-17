package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_public_ip"
)

func PublicIpFromTfToHttp(tf resource_public_ip.PublicIpModel) *api.PublicIpSchema {
	return &api.PublicIpSchema{}
}

func PublicIpFromHttpToTf(http *api.PublicIpSchema) resource_public_ip.PublicIpModel {
	return resource_public_ip.PublicIpModel{}
}

func PublicIpFromTfToCreateRequest(tf resource_public_ip.PublicIpModel) api.CreatePublicIpJSONRequestBody {
	return api.CreatePublicIpJSONRequestBody{}
}
