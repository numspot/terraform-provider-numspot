package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccInternetServiceResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testInternetServiceConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_internet_service.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_internet_service.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_internet_service.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testInternetServiceConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_internet_service.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_internet_service.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}
func testInternetServiceConfig_Create() string {
	return `resource "numspot_internet_service" "test" {
  			}`
}
func testInternetServiceConfig_Update() string {
		return `resource "numspot_internet_service" "test" {
    			}`
}
