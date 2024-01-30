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
	netIpRange := "10.101.0.0/16"
	subnetIpRange := "10.101.1.0/24"
	updatedIpRange := "10.101.2.0/24"

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
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
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
resource "numspot_net" "main" {
	ip_range = %[1]q
}

resource "numspot_subnet" "test" {
	net_id 		= numspot_net.main.id
	ip_range 	= %[2]q
}`, netIpRange, subnetIpRange)
}

func TestAccSubnetResource_MapPublicIp(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	// Required
	netIpRange := "10.101.0.0/16"
	subnetIpRange := "10.101.1.0/24"
	updatedIpRange := "10.101.2.0/24"

	// Computed
	subnetId := ""

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSubnetConfig_MapPublicIp(netIpRange, subnetIpRange),
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
				Config: testSubnetConfig_MapPublicIp(netIpRange, updatedIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "ip_range", updatedIpRange),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
						require.NotEqual(t, v, subnetId)
						return nil
					}),
				),
			},
		},
	})
}

func testSubnetConfig_MapPublicIp(netIpRange, subnetIpRange string) string {
	return fmt.Sprintf(`
resource "numspot_net" "main" {
	ip_range = %[1]q
}

resource "numspot_subnet" "test" {
	net_id 		= numspot_net.main.id
	ip_range 	= %[2]q
	map_public_ip_on_launch = true
}`, netIpRange, subnetIpRange)
}
