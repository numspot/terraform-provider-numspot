package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRouteTablesDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchRouteTableConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_route_table.testdata", "route_tables.#", "1"),
				),
			},
		},
	})
}

func fetchRouteTableConfig() string {
	return `
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.net.id
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.net.id
  subnet_id = numspot_subnet.subnet.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

data "numspot_route_table" "testdata" {
  ids        = [numspot_route_table.test.id]
  depends_on = [numspot_route_table.test]
}

`
}
