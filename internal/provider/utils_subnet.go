package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_subnet"
)

func SubnetFromTfToHttp(tf resource_subnet.SubnetModel) *api.SubnetSchema {
	return &api.SubnetSchema{}
}

func SubnetFromHttpToTf(http *api.SubnetSchema) resource_subnet.SubnetModel {
	return resource_subnet.SubnetModel{}
}

func SubnetFromTfToCreateRequest(tf resource_subnet.SubnetModel) api.CreateSubnetJSONRequestBody {
	return api.CreateSubnetJSONRequestBody{}
}
