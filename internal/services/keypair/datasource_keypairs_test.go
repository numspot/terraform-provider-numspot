//go:build acc

package keypair_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccKeypairDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	name := "key-pair-name-test-terraform"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchKeypairConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_keypairs.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_keypairs.testdata", "items.*", map[string]string{
						"name": name,
					}),
				),
			},
		},
	})
}

func fetchKeypairConfig(name string) string {
	return fmt.Sprintf(`
resource "numspot_keypair" "test" {
  name = %[1]q
}

data "numspot_keypairs" "testdata" {
  keypair_names = [numspot_keypair.test.name]
}
`, name)
}
