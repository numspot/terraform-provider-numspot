package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccSnapshotsDatasource(t *testing.T) {
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
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_snapshots.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_snapshots.testdata", "items.*", map[string]string{
						"id":        acctest.PAIR_PREFIX + "numspot_snapshot.test.id",
						"volume_id": acctest.PAIR_PREFIX + "numspot_volume.test.id",
					}),
				),
			},
		},
	})
}
