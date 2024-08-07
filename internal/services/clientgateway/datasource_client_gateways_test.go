//go:build acc

package clientgateway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccClientGatewaysDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	connectionType := "ipsec.1"
	publicIp := "192.0.2.0"
	bgpAsn := "65000"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchClientGatewayConfig(connectionType, publicIp, bgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_client_gateways.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_client_gateways.testdata", "items.*", map[string]string{
						"id":              provider.PAIR_PREFIX + "numspot_client_gateway.test.id",
						"connection_type": connectionType,
						"public_ip":       publicIp,
						"bgp_asn":         bgpAsn,
					}),
				),
			},
		},
	})
}

func fetchClientGatewayConfig(connectionType, publicIp, bgpAsn string) string {
	return fmt.Sprintf(`
resource "numspot_client_gateway" "test" {
  connection_type = %[1]q
  public_ip       = %[2]q
  bgp_asn         = %[3]s
}

data "numspot_client_gateways" "testdata" {
  ids = [numspot_client_gateway.test.id]

}
`, connectionType, publicIp, bgpAsn)
}
