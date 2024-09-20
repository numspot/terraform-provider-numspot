package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccKeypairDatasource(t *testing.T) {
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
resource "numspot_keypair" "test" {
  name = "key-pair-name-test-terraform"
}

data "numspot_keypairs" "testdata" {
  keypair_names = [numspot_keypair.test.name]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_keypairs.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_keypairs.testdata", "items.*", map[string]string{
						"name": "key-pair-name-test-terraform",
					}),
				),
			},
		},
	})
}
