package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccRouteTableResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testRouteTableConfig_Create(),
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
				ResourceName:      "numspot_route_table.test",
				ImportState:       true,
				ImportStateVerify: true,
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
func testRouteTableConfig_Create() string {
	return `resource "numspot_route_table" "test" {
  			}`
}
func testRouteTableConfig_Update() string {
		return `resource "numspot_route_table" "test" {
    			}`
}
