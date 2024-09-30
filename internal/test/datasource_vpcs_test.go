package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVPCsDatasource_Basic(t *testing.T) {
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
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

data "numspot_vpcs" "testdata" {
  ids = [numspot_vpc.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpcs.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vpcs.testdata", "items.*", map[string]string{
						"id":       acctest.PAIR_PREFIX + "numspot_vpc.test.id",
						"ip_range": "10.101.0.0/16",
					}),
				),
			},
		},
	})
}
