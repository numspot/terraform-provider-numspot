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
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = "65000"
}

data "numspot_client_gateways" "testdata" {
  ids = [numspot_client_gateway.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_client_gateways.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_client_gateways.testdata", "items.*", map[string]string{
						"id":              acctest.PAIR_PREFIX + "numspot_client_gateway.test.id",
						"connection_type": "ipsec.1",
						"public_ip":       "192.0.2.0",
						"bgp_asn":         "65000",
					}),
				),
			},
		},
	})
}
