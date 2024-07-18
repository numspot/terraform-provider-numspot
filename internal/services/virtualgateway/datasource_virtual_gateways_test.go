//go:build acc

package virtualgateway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccVirtualGatewaysDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories
	connection_type := "ipsec.1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVirtualGatewayConfig(connection_type),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_virtual_gateways.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_virtual_gateways.testdata", "items.*", map[string]string{
						"id":              provider.PAIR_PREFIX + "numspot_virtual_gateway.test.id",
						"connection_type": connection_type,
					}),
				),
			},
		},
	})
}

func fetchVirtualGatewayConfig(connection_type string) string {
	return fmt.Sprintf(`
resource "numspot_virtual_gateway" "test" {
  connection_type = %[1]q
}

data "numspot_virtual_gateways" "testdata" {
  ids = [numspot_virtual_gateway.test.id]
}
`, connection_type)
}
