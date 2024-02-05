package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nic"
)

func NicFromTfToHttp(tf *resource_nic.NicModel) *api.NicSchema {
	return &api.NicSchema{}
}

func NicFromHttpToTf(http *api.NicSchema) resource_nic.NicModel {
	return resource_nic.NicModel{}
}

func NicFromTfToCreateRequest(tf *resource_nic.NicModel) api.CreateNicJSONRequestBody {
	return api.CreateNicJSONRequestBody{}
}
