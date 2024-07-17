package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
)

func TestAccSnapshotsDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchSnapshotConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_snapshots.testdata", "items.#", "1"),
					utils_acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_snapshots.testdata", "items.*", map[string]string{
						"id":        utils_acctest.PAIR_PREFIX + "numspot_snapshot.test.id",
						"volume_id": utils_acctest.PAIR_PREFIX + "numspot_volume.test.id",
					}),
				),
			},
		},
	})
}

func fetchSnapshotConfig() string {
	return `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id = numspot_volume.test.id
}

data "numspot_snapshots" "testdata" {
  ids = [numspot_snapshot.test.id]
}
`
}
