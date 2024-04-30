//go:build acc

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInternetGatewaysDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchInternetGatewaysConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_internet_gateways.testdata", "internet_gateways.#", "1"),
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
