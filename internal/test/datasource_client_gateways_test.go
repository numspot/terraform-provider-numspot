package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccClientGatewaysDatasource(t *testing.T) {
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
resource "numspot_client_gateway" "terraform-client-gateway-acctest" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
  tags = [{
    key   = "name"
    value = "terraform-client-gateway-acctest"
  }]
}

data "numspot_client_gateways" "datasource-client-gateways-acctest" {
  ids = [numspot_client_gateway.terraform-client-gateway-acctest.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_client_gateways.datasource-client-gateways-acctest", "items.#", "1"),
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
