//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// optional : routes, vpn options
// dependencies : virtual gw, client gw,
//
// 1 - Create vpn connection (no routes & vpn options)
// 2 - import
// 3 - Update vpn connection (no routes & vpn options)
// 4 - Add vpn options (no routes)
// 5 - Replace vpn connection (no routes & vpn options)
// 6 - Update vpn options from vpn connection (no routes)
// 7 - Recreate vpn connection (no routes)
//
// 8 - Update vpn connection with 2 routes
// 9 - Update vpn connection with 0 routes
// 10- Update vpn connection with 1 routes
// 11- Recreate vpn connection
//
// 12- Associate vpn connection to new virtual gateway and client gateway
// 13- Associate vpn connection to new virtual gateway and client gateway (with deletion of old ones)
func TestAccVpnConnectionResource(t *testing.T) {
	// Skipped: backend connectivity slowness (503/EOF on POST /vpnConnections).
	// To be re-recorded once the IaaS team fixes the backend.
	t.Skip()
	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create vpn connection (no routes & vpn options)
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 12345
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  static_routes_only = "true"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", "true"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			{ // 2 - ImportState testing
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{ // 3 - Update vpn connection with 2 routes
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 12345
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  routes = [
    { destination_ip_range = "10.12.0.0/16" },
    { destination_ip_range = "10.15.0.0/16" }
  ]
  static_routes_only = "true"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", "true"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "10.12.0.0/16",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "10.15.0.0/16",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 4 - Update vpn connection with 0 routes
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 12345
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  routes = [
  ]
  static_routes_only = "true"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", "true"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "0"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 5 - Update vpn connection with 1 routes
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 12345
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  routes = [
    { destination_ip_range = "10.50.0.0/16" }
  ]
  static_routes_only = "true"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", "true"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "10.50.0.0/16",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 6 - Associate vpn connection to new virtual gateway and client gateway
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 12345
}

resource "numspot_client_gateway" "test_new" {
  connection_type = "ipsec.1"
  public_ip       = "192.10.2.0"
  bgp_asn         = 123456
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

resource "numspot_virtual_gateway" "test_new" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test_new.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test_new.id
  routes = [
    { destination_ip_range = "10.50.0.0/16" }
  ]
  static_routes_only = "true"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", "true"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "10.50.0.0/16",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test_new", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
