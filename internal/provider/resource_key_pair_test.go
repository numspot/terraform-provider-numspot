package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccKeyPairResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testKeyPairConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_key_pair.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_key_pair.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_key_pair.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testKeyPairConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_key_pair.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_key_pair.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testKeyPairConfig_Create() string {
	return fmt.Sprintf(`resource "numspot_key_pair" "test" {
  			}`)
}

func testKeyPairConfig_Update() string {
	return `resource "numspot_key_pair" "test" {
    			}`
}
