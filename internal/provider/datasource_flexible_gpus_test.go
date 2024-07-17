//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
)

func TestAccFlexibleGpusDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	model_name := "nvidia-a100-80"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchFlexibleGpusConfig(model_name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_flexible_gpus.testdata", "items.#", "1"),
					utils_acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_flexible_gpus.testdata", "items.*", map[string]string{
						"id":         utils_acctest.PAIR_PREFIX + "numspot_flexible_gpu.test.id",
						"model_name": model_name,
					}),
				),
			},
		},
	})
}

func fetchFlexibleGpusConfig(model_name string) string {
	return fmt.Sprintf(`
resource "numspot_flexible_gpu" "test" {
  model_name             = %[1]q
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

data "numspot_flexible_gpus" "testdata" {
  ids        = [numspot_flexible_gpu.test.id]
  depends_on = [numspot_flexible_gpu.test]
}
`, model_name)
}
