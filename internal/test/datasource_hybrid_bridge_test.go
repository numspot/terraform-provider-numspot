//go:build acc

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccHybridBridgeDatasource(t *testing.T) {
	managedServiceID := os.Getenv("NUMSPOT_TEST_HYBRID_BRIDGE_MANAGED_SERVICE_ID")
	if managedServiceID == "" {
		t.Skip("NUMSPOT_TEST_HYBRID_BRIDGE_MANAGED_SERVICE_ID is not set")
	}

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
				Config: fmt.Sprintf(`
resource "numspot_vpc" "vpc-source-hybrid-bridge-datasource-acctest" {
  ip_range = "10.101.1.0/24"
  tags = [{
    key   = "name"
    value = "terraform-hybrid-bridge-acctest"
  }]
}

resource "numspot_hybrid_bridge" "resource-hybrid-bridge-datasource-acctest" {
  managed_service_id = %q
  vpc_id             = numspot_vpc.vpc.id
}

data "numspot_hybrid_bridges" "datasource-hybrid-bridge-acctest" {
  depends_on = [numspot_hybrid_bridge.resource-hybrid-bridge-datasource-acctest]
}

output "selected_bridge" {
  value = data.numspot_hybrid_bridges.datasource-hybrid-bridge-acctest
}
`, managedServiceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_hybrid_bridges.datasource-hybrid-bridge-acctest", "items.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.numspot_hybrid_bridges.datasource-hybrid-bridge-acctest",
						"items.*",
						map[string]string{
							"source_ip_range":      "10.101.1.0/24",
							"destination_ip_range": "10.101.2.0/24",
						},
					),
				),
			},
		},
	})
}
