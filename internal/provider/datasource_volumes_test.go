//go:build acc

package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
)

func TestAccVolumesDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	volumeType := "standard"
	volumeSize := 11
	volumeAZ := "cloudgouv-eu-west-1a"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVolumesConfigById(volumeType, volumeSize, volumeAZ),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_volumes.testdata", "items.#", "1"),
					utils_acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_volumes.testdata", "items.*", map[string]string{
						"id":                     utils_acctest.PAIR_PREFIX + "numspot_volume.test.id",
						"type":                   volumeType,
						"size":                   strconv.Itoa(volumeSize),
						"availability_zone_name": volumeAZ,
					}),
				),
			},
		},
	})
}

func fetchVolumesConfigById(volumeType string, volumeSize int, volumeAZ string) string {
	return fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = %[1]q
  size                   = %[2]d
  availability_zone_name = %[3]q
}

data "numspot_volumes" "testdata" {
  ids = [numspot_volume.test.id]
}
`, volumeType, volumeSize, volumeAZ)
}
