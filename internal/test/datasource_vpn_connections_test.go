//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccVpnConnectionDatasource(t *testing.T) {
	t.Skip()
	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.0.0"
  bgp_asn         = 123456
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "terraform-vpn-connection-acctest" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  static_routes_only = false
}

data "numspot_vpn_connections" "testdata" {
  depends_on = [numspot_vpn_connection.terraform-vpn-connection-acctest]
}

output "selected_cg" {
  value = try([for c in data.numspot_vpn_connections.testdata.items : c if c.id == numspot_vpn_connection.terraform-vpn-connection-acctest.id][0], null)
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpn_connections.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vpn_connections.testdata", "items.*", map[string]string{
						"id":                 acctest.PairPrefix + "numspot_vpn_connection.terraform-vpn-connection-acctest.id",
						"client_gateway_id":  acctest.PairPrefix + "numspot_client_gateway.test.id",
						"virtual_gateway_id": acctest.PairPrefix + "numspot_virtual_gateway.test.id",
					}),
				),
			},
		},
	})
}
