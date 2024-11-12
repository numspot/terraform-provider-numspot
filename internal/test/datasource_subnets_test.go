package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccSubnetsDatasource(t *testing.T) {
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
data "numspot_subnets" "testdata" {
  vpc_ids    = [numspot_vpc.main.id]
  depends_on = [numspot_subnet.test]
}
resource "numspot_vpc" "main" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.main.id
  ip_range = "10.101.1.0/24"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_subnets.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_subnets.testdata", "items.*", map[string]string{
						"id":       acctest.PairPrefix + "numspot_subnet.test.id",
						"ip_range": "10.101.1.0/24",
					}),
				),
			},
		},
	})
}
