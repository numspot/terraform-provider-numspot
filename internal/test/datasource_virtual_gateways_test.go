//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccVirtualGatewaysDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t)
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

resource "numspot_virtual_gateway" "terraform-virtual-gateway-acctest" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
}

data "numspot_virtual_gateways" "datasource-virtual-gateways-acctest" {
  depends_on = [numspot_virtual_gateway.terraform-virtual-gateway-acctest]
}

output "selected_vg" {
  value = try([for c in data.numspot_virtual_gateways.datasource-virtual-gateways-acctest.items : c if c.id == numspot_virtual_gateway.terraform-virtual-gateway-acctest.id][0], null)
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.TestCheckTypeSetElemNestedAttrsWithPair(
						"data.numspot_virtual_gateways.datasource-virtual-gateways-acctest",
						"items.*",
						map[string]string{
							"id":              acctest.PairPrefix + "numspot_virtual_gateway.terraform-virtual-gateway-acctest.id",
							"connection_type": "ipsec.1",
						},
					),
				),
			},
		},
	})
}
