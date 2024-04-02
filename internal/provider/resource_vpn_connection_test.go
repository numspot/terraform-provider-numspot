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
				ImportStateVerifyIgnore: []string{},
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
		public_ip = "192.0.2.0"
		bgp_asn = 65000
	}

    resource "numspot_virtual_gateway" "test" {
		connection_type = "ipsec.1"
	}

	resource "numspot_vpn_connection" "test" {
		client_gateway_id = numspot_client_gateway.test.id
		connection_type = "ipsec.1"
		virtual_gateway_id = numspot_virtual_gateway.test.id
		static_routes_only = %[1]q
	}

	`, static_route_only)
}

func TestAccVpnConnectionResource_UpdateGateways(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	var (
		vpn_connection_id  string
		client_gateway_id  string
		virtual_gateway_id string
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVpnConnectionConfig_UpdateGateways("client_gateway_before", "virtual_gateway_before"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						vpn_connection_id = v
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "client_gateway_id", func(v string) error {
						if v == "" {
							return errors.New("client_gateway_id should not be empty")
						}
						client_gateway_id = v
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "virtual_gateway_id", func(v string) error {
						if v == "" {
							return errors.New("virtual_gateway_id should not be empty")
						}
						virtual_gateway_id = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testVpnConnectionConfig_UpdateGateways("client_gateway_after", "virtual_gateway_after"),
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
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "client_gateway_id", func(v string) error {
						if v == "" {
							return errors.New("client_gateway_id should not be empty")
						}
						if client_gateway_id == v {
							return errors.New("client_gateway_id should be different after Update")
						}

						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "virtual_gateway_id", func(v string) error {
						if v == "" {
							return errors.New("virtual_gateway_id should not be empty")
						}
						if virtual_gateway_id == v {
							return errors.New("virtual_gateway_id should be different after Update")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testVpnConnectionConfig_UpdateGateways(clientGatewayName string, virtualGatewayName string) string {
	return fmt.Sprintf(`
	resource "numspot_client_gateway" "client_gateway_before" {
		connection_type = "ipsec.1"
		public_ip = "192.0.2.0"
		bgp_asn = 65000
	}

	resource "numspot_virtual_gateway" "virtual_gateway_before" {
		connection_type = "ipsec.1"
	}

	resource "numspot_client_gateway" "client_gateway_after" {
		connection_type = "ipsec.1"
		public_ip = "192.0.2.1"
		bgp_asn = 65001
	}

	resource "numspot_virtual_gateway" "virtual_gateway_after" {
		connection_type = "ipsec.1"
	}

	resource "numspot_vpn_connection" "test" {
		client_gateway_id = numspot_client_gateway.%[1]s.id
		connection_type = "ipsec.1"
		virtual_gateway_id = numspot_virtual_gateway.%[2]s.id
		static_routes_only = "false"
	}
	`, clientGatewayName, virtualGatewayName)
}
