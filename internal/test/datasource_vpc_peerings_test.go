package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVpcPeeringsDatasource(t *testing.T) {
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
resource "numspot_vpc" "source" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "accepter" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter.id
  source_vpc_id   = numspot_vpc.source.id
}

data "numspot_vpc_peerings" "testdata" {
  ids = [numspot_vpc_peering.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpc_peerings.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vpc_peerings.testdata", "items.*", map[string]string{
						"id": acctest.PairPrefix + "numspot_vpc_peering.test.id",
					}),
				),
			},
		},
	})
}
