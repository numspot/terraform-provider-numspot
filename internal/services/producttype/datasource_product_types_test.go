//go:build acc

package producttype_test

/*

 PRODUCT TYPES are not handled for now


import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProductTypesDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchProductTypesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_product_types.testdata", "items.#", "1"),
				),
			},
		},
	})
}

func fetchProductTypesConfig() string {
	return `
data "numspot_product_types" "testdata" {
  ids = ["0001"]
}
`
}
*/
