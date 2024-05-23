//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFlexibleGpuResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	flexibleGpuModelName := "nvidia-a100-80"
	flexibleGpuGeneration := "v6"
	flexibleGpuAZ := "cloudgouv-eu-west-1a"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testFlexibleGpuConfig(flexibleGpuModelName, flexibleGpuGeneration, flexibleGpuAZ),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", flexibleGpuModelName),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", flexibleGpuGeneration),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", flexibleGpuAZ),
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

func testFlexibleGpuConfig(modelName, generation, az string) string {
	return fmt.Sprintf(`
resource "numspot_flexible_gpu" "test" {
  model_name             = %[1]q
  generation             = %[2]q
  availability_zone_name = %[3]q
}`, modelName, generation, az)
}
