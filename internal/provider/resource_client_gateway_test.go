package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccClientGatewayResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	connectionType := "ipsec.1"
	public_ip := "192.0.2.0"
	bgpAsn := 65000

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testClientGatewayConfig_Create(connectionType, public_ip, bgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "connection_type", connectionType),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", public_ip),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "bgp_asn", func(v string) error {
						require.NotEmpty(t, v)
						require.Equal(t, bgpAsn, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_client_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testClientGatewayConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testClientGatewayConfig_Create(connectionType, publicIp string, bgpAsn int) string {
	return fmt.Sprintf(`resource "numspot_client_gateway" "test" {
		connection_type = %[1]q
		public_ip = %[2]q
		bgp_asn = %d
	}`, connectionType, publicIp, bgpAsn)
}

func testClientGatewayConfig_Update() string {
	return `resource "numspot_client_gateway" "test" {
		
	}`
}
