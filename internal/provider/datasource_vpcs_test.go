//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVPCsDatasource_Basic(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	ipRange := "10.101.0.0/16"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCsDatasourceConfig_Basic(ipRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpcs.test", "vpcs.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_vpcs.test", "vpcs.0.ip_range", ipRange),
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
