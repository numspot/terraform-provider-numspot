//go:build acc

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVpnConnectionResource_UpdateStaticRouteOnly(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	static_route_only := "false"
	static_route_only_updated := "true"
	var vpn_connection_id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVpnConnectionConfig_UpdateStaticRouteOnly(static_route_only),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", static_route_only),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						vpn_connection_id = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			// Update testing
			{
				Config: testVpnConnectionConfig_UpdateStaticRouteOnly(static_route_only_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", static_route_only_updated),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						if vpn_connection_id == v {
							return errors.New("Id should be different after Update")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testVpnConnectionConfig_UpdateStaticRouteOnly(static_route_only string) string {
	return fmt.Sprintf(`


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
  static_routes_only = %[1]q
}`, static_route_only)
}

func TestAccVpnConnectionResource_WithRoute(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	destinationIPRange := "10.12.0.0/16"
	updatedRoutes := []string{"192.0.2.0/24", "192.168.255.0/24"}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVpnConnectionConfig_WithSingleRoute(destinationIPRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "1"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.0.destination_ip_range", destinationIPRange),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			{
				Config: testVpnConnectionConfig_WithMultiRoutes(updatedRoutes),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", "2"),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.0.destination_ip_range", updatedRoutes[0]),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.1.destination_ip_range", updatedRoutes[1]),
				),
			},
		},
	})
}

func testVpnConnectionConfig_WithSingleRoute(ipRange string) string {
	return fmt.Sprintf(`


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
  static_routes_only = "true"
  routes = [
    {
      destination_ip_range = %[1]q
    }
  ]
}`, ipRange)
}

func testVpnConnectionConfig_WithMultiRoutes(routes []string) string {
	return fmt.Sprintf(`


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
  static_routes_only = "true"
  routes = [
    {
      destination_ip_range = %[1]q
    },
    {
      destination_ip_range = %[2]q
    }
  ]
}`, routes[0], routes[1])
}

func TestAccVpnConnectionResource_UpdateTunnelEndpoints(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	var vpn_connection_id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVpnConnectionConfig_UpdateTunnelEndpoints("client_gateway_before", "virtual_gateway_before"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						vpn_connection_id = v
						return nil
					}),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_vpn_connection.test", "client_gateway_id",
						"numspot_client_gateway.client_gateway_before", "id"),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_vpn_connection.test", "virtual_gateway_id",
						"numspot_virtual_gateway.virtual_gateway_before", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			// Update testing
			{
				Config: testVpnConnectionConfig_UpdateTunnelEndpoints("client_gateway_after", "virtual_gateway_after"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}

						if vpn_connection_id == v {
							return errors.New("Id should be different after Update")
						}

						return nil
					}),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_vpn_connection.test", "client_gateway_id",
						"numspot_client_gateway.client_gateway_after", "id"),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_vpn_connection.test", "virtual_gateway_id",
						"numspot_virtual_gateway.virtual_gateway_after", "id"),
				),
			},
		},
	})
}

func testVpnConnectionConfig_UpdateTunnelEndpoints(clientGatewayName string, virtualGatewayName string) string {
	return fmt.Sprintf(`
resource "numspot_client_gateway" "client_gateway_before" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_virtual_gateway" "virtual_gateway_before" {
  connection_type = "ipsec.1"
}

resource "numspot_client_gateway" "client_gateway_after" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.1"
  bgp_asn         = 65001
}

resource "numspot_virtual_gateway" "virtual_gateway_after" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.%[1]s.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.%[2]s.id
  static_routes_only = "false"
}
	`, clientGatewayName, virtualGatewayName)
}

func TestAccVpnConnectionResource_UpdateVPNOptions(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	tunnelInsideIpRange := "169.254.254.22/30"
	presharedKey := "sample key !"
	tunnelInsideIpRangeUpdated := "169.254.254.20/30"
	presharedKeyUpdated := "another key !"
	var vpn_connection_id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVpnConnectionConfig_UpdateVPNOptions(tunnelInsideIpRange, presharedKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						vpn_connection_id = v
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.phase2options.pre_shared_key", presharedKey),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.tunnel_inside_ip_range", tunnelInsideIpRange),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			// Update testing
			{
				Config: testVpnConnectionConfig_UpdateVPNOptions(tunnelInsideIpRangeUpdated, presharedKeyUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}

						if vpn_connection_id != v {
							return errors.New("Id should be identical after Update")
						}

						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.phase2options.pre_shared_key", presharedKeyUpdated),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.tunnel_inside_ip_range", tunnelInsideIpRangeUpdated),
				),
			},
		},
	})
}

func testVpnConnectionConfig_UpdateVPNOptions(tunnelInsideIPRange, presharedKey string) string {
	return fmt.Sprintf(`
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
  static_routes_only = "false"
  vpn_options = {
    phase2options = {
      pre_shared_key = %[2]q
    }
    tunnel_inside_ip_range = %[1]q
  }
}
`, tunnelInsideIPRange, presharedKey)
}

func TestAccVpnConnectionResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	static_route_only := "false"
	var vpn_connection_id string

	tagKey := "Name"
	tagValue := "Terraform-Test-Volume"
	tagValueUpdated := tagValue + "-Updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVpnConnectionConfig_Tags(static_route_only, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", static_route_only),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("id should not be empty")
						}
						vpn_connection_id = v
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			// Update testing
			{
				Config: testVpnConnectionConfig_Tags(static_route_only, tagKey, tagValueUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", static_route_only),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						if vpn_connection_id != v {
							return errors.New("id should not be different after Update")
						}
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.0.value", tagValueUpdated),
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testVpnConnectionConfig_Tags(static_route_only, tagKey, tagValue string) string {
	return fmt.Sprintf(`
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
  static_routes_only = %[1]q

  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, static_route_only, tagKey, tagValue)
}
