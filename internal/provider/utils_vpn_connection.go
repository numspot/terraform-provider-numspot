package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpn_connection"
)

func VpnConnectionFromTfToHttp(tf resource_vpn_connection.VpnConnectionModel) *api.VpnConnectionSchema {
	return &api.VpnConnectionSchema{}
}

func VpnConnectionFromHttpToTf(http *api.VpnConnectionSchema) resource_vpn_connection.VpnConnectionModel {
	return resource_vpn_connection.VpnConnectionModel{}
}

func VpnConnectionFromTfToCreateRequest(tf resource_vpn_connection.VpnConnectionModel) api.CreateVpnConnectionJSONRequestBody {
	return api.CreateVpnConnectionJSONRequestBody{}
}
