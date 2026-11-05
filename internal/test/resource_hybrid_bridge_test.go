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

func TestAccHybridBridgeResource(t *testing.T) {
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
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_hybrid_bridge" "resource-hybrid-bridge-resource-acctest" {
  managed_service_id = %q
  vpc_id             = numspot_vpc.vpc.id
}
`, managedServiceID),
				Check: resource.ComposeTestCheckFunc(
					// resource.TestCheckResourceAttrPair("numspot_hybrid_bridge.resource-hybrid-bridge-resource-acctest", "managed_service_id", "numspot_hybrid_bridge.resource-hybrid-bridge-resource-acctest", "managed_service_id"),
					resource.TestCheckResourceAttrPair(
						"numspot_hybrid_bridge.resource-hybrid-bridge-resource-acctest",
						"vpc_id",
						"numspot_vpc.vpc",
						"id",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"numspot_hybrid_bridge.resource-hybrid-bridge-resource-acctest",
						"route.*",
						map[string]string{
							"destination_ip_range": "10.0.0.0/16",
						},
					),
				),
			},
		},
	})
}
