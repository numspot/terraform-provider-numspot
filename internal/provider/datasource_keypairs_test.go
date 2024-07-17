package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
)

func TestAccKeypairDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	name := "key-pair-name-test-terraform"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchKeypairConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_keypairs.testdata", "items.#", "1"),
					utils_acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_keypairs.testdata", "items.*", map[string]string{
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
  names = [numspot_keypair.test.name]
}
`, name)
}
