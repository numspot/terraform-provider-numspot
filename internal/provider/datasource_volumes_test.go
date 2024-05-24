//go:build acc

package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumesDatasource(t *testing.T) {
	t.Parallel()
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
					resource.TestCheckResourceAttr("data.numspot_volumes.datasource_test", "items.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_volumes.datasource_test", "items.0.type", volumeType),
					resource.TestCheckResourceAttr("data.numspot_volumes.datasource_test", "items.0.size", strconv.Itoa(volumeSize)),
					resource.TestCheckResourceAttr("data.numspot_volumes.datasource_test", "items.0.availability_zone_name", volumeAZ),
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

data "numspot_volumes" "datasource_test" {
  ids = [numspot_volume.test.id]
}
`, volumeType, volumeSize, volumeAZ)
}
