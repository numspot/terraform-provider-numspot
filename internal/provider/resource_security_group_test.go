package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccSecurityGroupResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_security_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSecurityGroupConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}
func testSecurityGroupConfig_Create() string {
	return `resource "numspot_security_group" "test" {
  			}`
}
func testSecurityGroupConfig_Update() string {
		return `resource "numspot_security_group" "test" {
    			}`
}
