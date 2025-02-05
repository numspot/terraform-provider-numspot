package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccNatGatewaysDatasource(t *testing.T) {
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
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_nat_gateways.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_nat_gateways.testdata", "items.*", map[string]string{
						"id":        acctest.PairPrefix + "numspot_nat_gateway.test.id",
						"subnet_id": acctest.PairPrefix + "numspot_subnet.test.id",
					}),
				),
			},
		},
	})
}
