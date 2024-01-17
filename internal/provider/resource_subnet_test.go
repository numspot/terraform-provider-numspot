package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccSubnetResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSubnetConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_subnet.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSubnetConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}
func testSubnetConfig_Create() string {
	return `resource "numspot_subnet" "test" {
  			}`
}
func testSubnetConfig_Update() string {
		return `resource "numspot_subnet" "test" {
    			}`
}
