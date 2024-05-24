package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVpnConnectionDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVpnConnectionConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpn_connections.testdata", "items.#", "1"),
				),
			},
		},
	})
}

func fetchVpnConnectionConfig() string {
	return `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  static_routes_only = false
}

data "numspot_vpn_connections" "testdata" {
  ids        = [numspot_vpn_connection.test.id]
  depends_on = [numspot_vpn_connection.test]
}
`
}
