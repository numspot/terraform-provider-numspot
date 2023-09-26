package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNicResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNicConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nic.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNicConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nic.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testNicConfig_Create() string {
	return fmt.Sprintf(`resource "numspot_nic" "test" {
  			}`)
}

func testNicConfig_Update() string {
	return `resource "numspot_nic" "test" {
    			}`
}
