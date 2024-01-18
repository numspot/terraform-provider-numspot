package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccImageResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testImageConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_image.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testImageConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_image.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testImageConfig_Create() string {
	return fmt.Sprintf(`resource "numspot_image" "test" {
  			}`)
}

func testImageConfig_Update() string {
	return `resource "numspot_image" "test" {
    			}`
}
