package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFlexibleGpuResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testFlexibleGpuConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_flexible_gpu.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testFlexibleGpuConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testFlexibleGpuConfig_Create() string {
	return fmt.Sprintf(`resource "numspot_flexible_gpu" "test" {
  			}`)
}

func testFlexibleGpuConfig_Update() string {
	return `resource "numspot_flexible_gpu" "test" {
    			}`
}
