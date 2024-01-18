package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccRouteTableResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	// Required
	netIpRange := "10.0.0.0/16"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testRouteTableConfig(netIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_route_table.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testRouteTableConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testRouteTableConfig(netIpRange string) string {
	return fmt.Sprintf(`
resource "net" "main" {
	ip_range = %[1]q
}

resource "numspot_route_table" "test" {
	net_id = net.main.id
}`, netIpRange)
}

func testRouteTableConfig_Update() string {
	return `resource "numspot_route_table" "test" {
    			}`
}
