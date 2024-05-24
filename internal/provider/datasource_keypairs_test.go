package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeypairDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchKeypairConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_keypairs.testdata", "items.#", "1"),
				),
			},
		},
	})
}

func fetchKeypairConfig() string {
	return `
resource "numspot_keypair" "test" {
  name = "key-pair-name"
}

data "numspot_keypairs" "testdata" {
  names      = [numspot_keypair.test.name]
  depends_on = [numspot_keypair.test]
}
`
}
