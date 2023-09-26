package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nat_service"
)

func NatServiceFromTfToHttp(tf resource_nat_service.NatServiceModel) *api.NatServiceSchema {
	return &api.NatServiceSchema{}
}

func NatServiceFromHttpToTf(http *api.NatServiceSchema) resource_nat_service.NatServiceModel {
	return resource_nat_service.NatServiceModel{}
}

func NatServiceFromTfToCreateRequest(tf resource_nat_service.NatServiceModel) api.CreateNatServiceJSONRequestBody {
	return api.CreateNatServiceJSONRequestBody{}
}
