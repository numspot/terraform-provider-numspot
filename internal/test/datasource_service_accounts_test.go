package test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccServiceAccountDatasource(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skipf("skipping %s test in CI", t.Name())
	}
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
resource "numspot_service_account" "test" {
  space_id = "67d97ad4-3005-48dc-a392-60a97ab5097c"
  name     = "terraform-service-account-test-datasource"
}

data "numspot_service_accounts" "testdata" {
  space_id            = numspot_service_account.test.space_id
  service_account_ids = [numspot_service_account.test.service_account_id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_service_accounts.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_service_accounts.testdata", "items.*", map[string]string{
						"id":   acctest.PAIR_PREFIX + "numspot_service_account.test.service_account_id",
						"name": "terraform-service-account-test-datasource",
					}),
				),
			},
		},
	})
}
