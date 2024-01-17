package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_client_gateway"
)

func ClientGatewayFromTfToHttp(tf resource_client_gateway.ClientGatewayModel) *api.ClientGatewaySchema {
	return &api.ClientGatewaySchema{}
}

func ClientGatewayFromHttpToTf(http *api.ClientGatewaySchema) resource_client_gateway.ClientGatewayModel {
	return resource_client_gateway.ClientGatewayModel{}
}

func ClientGatewayFromTfToCreateRequest(tf resource_client_gateway.ClientGatewayModel) api.CreateClientGatewayJSONRequestBody {
	return api.CreateClientGatewayJSONRequestBody{}
}
