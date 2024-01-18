package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccSubnetResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	// Required
	netIpRange := "10.0.0.0/16"
	subnetIpRange := "10.0.1.0/24"
	updatedIpRange := "10.0.2.0/24"

	// Computed
	subnetId := ""

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSubnetConfig(netIpRange, subnetIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "ip_range", subnetIpRange),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						subnetId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_subnet.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSubnetConfig(netIpRange, updatedIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "ip_range", updatedIpRange),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "field", func(v string) error {
						require.NotEqual(t, v, subnetId)
						return nil
					}),
				),
			},
		},
	})
}

func testSubnetConfig(netIpRange, subnetIpRange string) string {
	return fmt.Sprintf(`
resource "net" "main" {
	ip_range = %[1]q
}

resource "numspot_subnet" "test" {
	net_id 		= net.main.id
	ip_range 	= %[2]q
}`, netIpRange, subnetIpRange)
}
