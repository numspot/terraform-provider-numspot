package provider

import (
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
				Config: testFlexibleGpuConfig(),
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
			//{
			//	Config: testFlexibleGpuConfig(),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "field", "value"),
			//		resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "field", func(v string) error {
			//			return nil
			//		}),
			//	),
			//},
		},
	})
}

func testFlexibleGpuConfig() string {
	return `
resource "numspot_flexible_gpu" "test" {
	model_name 				= "nvidia-p100"
	generation 				= "v5"
	availability_zone_name 	= "eu-west-2a"
}`
}
