package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccPublicIpResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testPublicIpConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_public_ip.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_public_ip.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testPublicIpConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_public_ip.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}
func testPublicIpConfig_Create() string {
	return `resource "numspot_public_ip" "test" {
  			}`
}
func testPublicIpConfig_Update() string {
		return `resource "numspot_public_ip" "test" {
    			}`
}
