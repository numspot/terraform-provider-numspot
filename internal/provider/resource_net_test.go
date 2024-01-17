package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccNetResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_net.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_net.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_net.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}
func testNetConfig_Create() string {
	return `resource "numspot_net" "test" {
  			}`
}
func testNetConfig_Update() string {
		return `resource "numspot_net" "test" {
    			}`
}
