package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVpnConnectionResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create testing
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 12345
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  routes = [
    { destination_ip_range = "10.12.0.0/16" }
  ]
  static_routes_only = "true"
  tags = [
    {
      key   = "name"
      value = "VPN-Connection-Test"
    }
  ]
  vpn_options = {
    phase2options = {
      pre_shared_key = "sample key !"
    }
    tunnel_inside_ip_range = "169.254.254.22/30"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", "true"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.phase2options.pre_shared_key", "sample key !"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.tunnel_inside_ip_range", "169.254.254.22/30"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "VPN-Connection-Test",
					}),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "10.12.0.0/16",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 12345
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  routes = [
    { destination_ip_range = "192.0.2.0/24" },
    { destination_ip_range = "192.168.255.0/24" }
  ]
  static_routes_only = "true"
  tags = [
    {
      key   = "name"
      value = "VPN-Connection-Test-Updated"
    }
  ]
  vpn_options = {
    phase2options = {
      pre_shared_key = "another key !"
    }
    tunnel_inside_ip_range = "169.254.254.20/30"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", "true"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.phase2options.pre_shared_key", "another key !"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.tunnel_inside_ip_range", "169.254.254.20/30"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "VPN-Connection-Test-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "192.0.2.0/24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "192.168.255.0/24",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// <== If resource has required dependencies ==>
			// 4 - Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_client_gateway" "test_new" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 54321
}

resource "numspot_virtual_gateway" "test_new" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test_new.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test_new.id
  routes = [
    { destination_ip_range = "192.0.2.0/24" },
    { destination_ip_range = "192.168.255.0/24" }
  ]
  static_routes_only = "true"
  tags = [
    {
      key   = "name"
      value = "VPN-Connection-Test-Updated"
    }
  ]
  vpn_options = {
    phase2options = {
      pre_shared_key = "another key !"
    }
    tunnel_inside_ip_range = "169.254.254.20/30"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", "true"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.phase2options.pre_shared_key", "another key !"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.tunnel_inside_ip_range", "169.254.254.20/30"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "VPN-Connection-Test-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "192.0.2.0/24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
						"destination_ip_range": "192.168.255.0/24",
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
