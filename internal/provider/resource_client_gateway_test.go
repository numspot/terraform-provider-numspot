package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccClientGatewayResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testClientGatewayConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_client_gateway.test",
				ImportState:       true,
				ImportStateVerify: true,
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
func testClientGatewayConfig_Create() string {
	return `resource "numspot_client_gateway" "test" {
  			}`
}
func testClientGatewayConfig_Update() string {
		return `resource "numspot_client_gateway" "test" {
    			}`
}
