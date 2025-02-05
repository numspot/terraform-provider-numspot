package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccFlexibleGpusDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: `
resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1b"
}

data "numspot_flexible_gpus" "testdata" {
  ids = [numspot_flexible_gpu.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_flexible_gpus.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_flexible_gpus.testdata", "items.*", map[string]string{
						"id":         acctest.PairPrefix + "numspot_flexible_gpu.test.id",
						"model_name": "nvidia-a100-80",
					}),
				),
			},
		},
	})
}
