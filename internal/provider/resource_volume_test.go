package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccVolumeResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVolumeConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_volume.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testVolumeConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_volume.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}
func testVolumeConfig_Create() string {
	return `resource "numspot_volume" "test" {
  			}`
}
func testVolumeConfig_Update() string {
		return `resource "numspot_volume" "test" {
    			}`
}
