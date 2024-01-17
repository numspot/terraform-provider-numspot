package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccNatServiceResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNatServiceConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_service.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_nat_service.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_nat_service.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNatServiceConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_service.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_nat_service.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}
func testNatServiceConfig_Create() string {
	return `resource "numspot_nat_service" "test" {
  			}`
}
func testNatServiceConfig_Update() string {
		return `resource "numspot_nat_service" "test" {
    			}`
}
