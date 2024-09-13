package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVirtualGatewaysDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider
	connection_type := "ipsec.1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVirtualGatewayConfig(connection_type),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_virtual_gateways.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_virtual_gateways.testdata", "items.*", map[string]string{
						"id":              acctest.PAIR_PREFIX + "numspot_virtual_gateway.test.id",
						"connection_type": connection_type,
					}),
				),
			},
		},
	})
}

func fetchVirtualGatewayConfig(connection_type string) string {
	return fmt.Sprintf(`

resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = %[1]q
  vpc_id          = numspot_vpc.test.id
}

data "numspot_virtual_gateways" "testdata" {
  ids = [numspot_virtual_gateway.test.id]
}
`, connection_type)
}
