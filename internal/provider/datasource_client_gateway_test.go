package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccClientGatewaysDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchClientGatewayConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_client_gateway.testdata", "client_gateways.#", "1"),
				),
			},
		},
	})
}

func fetchClientGatewayConfig() string {
	return `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

data "numspot_client_gateway" "testdata" {
  ids        = [numspot_client_gateway.test.id]
  depends_on = [numspot_client_gateway.test]

}
`
}
