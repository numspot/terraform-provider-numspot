//go:build acc

package internetgateway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccInternetGatewaysDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchInternetGatewaysConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_internet_gateways.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_internet_gateways.testdata", "items.*", map[string]string{
						"id":     provider.PAIR_PREFIX + "numspot_internet_gateway.test.id",
						"vpc_id": provider.PAIR_PREFIX + "numspot_vpc.test.id",
					}),
				),
			},
		},
	})
}

func fetchInternetGatewaysConfig() string {
	return `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

data "numspot_internet_gateways" "testdata" {
  ids        = [numspot_internet_gateway.test.id]
  depends_on = [numspot_internet_gateway.test]
}
`
}
