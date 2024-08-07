//go:build acc

package natgateway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccNatGatewaysDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchNatGatewaysConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_nat_gateways.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_nat_gateways.testdata", "items.*", map[string]string{
						"id":        provider.PAIR_PREFIX + "numspot_nat_gateway.test.id",
						"subnet_id": provider.PAIR_PREFIX + "numspot_subnet.test.id",
					}),
				),
			},
		},
	})
}

func fetchNatGatewaysConfig() string {
	return `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test" {
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test.id
  public_ip_id = numspot_public_ip.test.id
  depends_on   = [numspot_route_table.test]
}

data "numspot_nat_gateways" "testdata" {
  subnet_ids = [numspot_subnet.test.id]
  depends_on = [numspot_nat_gateway.test]
}
`
}
