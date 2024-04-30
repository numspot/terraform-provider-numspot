package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVirtualGatewaysDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVirtualGatewayConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_virtual_gateways.testdata", "virtual_gateways.#", "1"),
				),
			},
		},
	})
}

func fetchVirtualGatewayConfig() string {
	return `
resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

data "numspot_virtual_gateways" "testdata" {
  ids        = [numspot_virtual_gateway.test.id]
  depends_on = [numspot_virtual_gateway.test]
}
`
}
