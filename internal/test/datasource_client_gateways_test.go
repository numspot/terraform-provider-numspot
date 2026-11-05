//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccClientGatewaysDatasource(t *testing.T) {
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
resource "numspot_client_gateway" "terraform-client-gateway-acctest" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

data "numspot_client_gateways" "datasource-client-gateways-acctest" {
  depends_on = [numspot_client_gateway.terraform-client-gateway-acctest]
}

output "selected_cg" {
  value = try([for c in data.numspot_client_gateways.datasource-client-gateways-acctest.items : c if c.id == numspot_client_gateway.terraform-client-gateway-acctest.id][0], null)
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.numspot_client_gateways.datasource-client-gateways-acctest", "items.*",
						map[string]string{
							"connection_type": "ipsec.1",
							"public_ip":       "192.0.2.0",
							"bgp_asn":         "65000",
						},
					),
				),
			},
		},
	})
}
