//go:build acc

package vpnconnection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccVpnConnectionDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVpnConnectionConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpn_connections.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vpn_connections.testdata", "items.*", map[string]string{
						"id":                 provider.PAIR_PREFIX + "numspot_vpn_connection.test.id",
						"client_gateway_id":  provider.PAIR_PREFIX + "numspot_client_gateway.test.id",
						"virtual_gateway_id": provider.PAIR_PREFIX + "numspot_virtual_gateway.test.id",
					}),
				),
			},
		},
	})
}

func fetchVpnConnectionConfig() string {
	return `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id

}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  static_routes_only = false
}

data "numspot_vpn_connections" "testdata" {
  ids = [numspot_vpn_connection.test.id]
}
`
}
