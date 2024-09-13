package test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVolumesDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

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
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_volumes.testdata", "items.*", map[string]string{
						"id":                     acctest.PAIR_PREFIX + "numspot_volume.test.id",
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
