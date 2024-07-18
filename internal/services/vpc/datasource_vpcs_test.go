//go:build acc

package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccVPCsDatasource_Basic(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories
	ipRange := "10.101.0.0/16"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCsDatasourceConfig_Basic(ipRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpcs.test", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vpcs.testdata", "items.*", map[string]string{
						"id":       provider.PAIR_PREFIX + "numspot_vpc.test.id",
						"ip_range": ipRange,
					}),
				),
			},
		},
	})
}

func testAccVPCsDatasourceConfig_Basic(ipRange string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = %[1]q
}

data "numspot_vpcs" "test" {
  ids = [numspot_vpc.test.id]
}`, ipRange)
}
