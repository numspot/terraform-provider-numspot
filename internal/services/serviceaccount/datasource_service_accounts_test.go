//go:build acc

package serviceaccount_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccServiceAccountDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	spaceID := "67d97ad4-3005-48dc-a392-60a97ab5097c"
	randint := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("terraform-service-account-test-%d", randint)
	// name := "My custom TF svc account"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchServiceAccountsConfig(spaceID, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_service_accounts.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_service_accounts.testdata", "items.*", map[string]string{
						"id":   provider.PAIR_PREFIX + "numspot_service_account.test.service_account_id",
						"name": name,
					}),
				),
			},
		},
	})
}

func fetchServiceAccountsConfig(spaceID, name string) string {
	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = %[2]q
}

data "numspot_service_accounts" "testdata" {
  space_id            = numspot_service_account.test.space_id
  service_account_ids = [numspot_service_account.test.service_account_id]
}
	`, spaceID, name)
}
