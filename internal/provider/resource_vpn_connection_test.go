package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccVpnConnectionResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVpnConnectionConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "field", func(v string) error {
						require.NotEmpty(t, v)
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
				Config: testVpnConnectionConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpn_connection.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testVpnConnectionConfig_Create() string {
	return `resource "numspot_vpn_connection" "test" {}`
}

func testVpnConnectionConfig_Update() string {
	return `resource "numspot_vpn_connection" "test" {
    			}`
}
