//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccRouteTableResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	netIpRange := "10.101.0.0/16"
	subnetIpRange := "10.101.1.0/24"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testRouteTableConfig(netIpRange, subnetIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_route_table.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testRouteTableConfig_Update(netIpRange, subnetIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// Update testing 2
			{
				Config: testRouteTableConfig(netIpRange, subnetIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// Update testing - Remove Subnet
			{
				Config: testRouteTableConfig_WithoutSubnet(netIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// Update testing - Re-add subnet
			{
				Config: testRouteTableConfig(netIpRange, subnetIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// Update testing - Unlink subnet without deleting it
			{
				Config: testRouteTableConfig_WithoutLinkSubnet(netIpRange, subnetIpRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
		},
	})
}

func testRouteTableConfig(netIpRange, subnetIpRange string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = %[2]q
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.net.id
  subnet_id = numspot_subnet.subnet.id
}`, netIpRange, subnetIpRange)
}

func testRouteTableConfig_WithoutSubnet(netIpRange string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_route_table" "test" {
  vpc_id = numspot_vpc.net.id
}`, netIpRange)
}

func testRouteTableConfig_WithoutLinkSubnet(netIpRange, subnetIpRange string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = %[2]q
}

resource "numspot_route_table" "test" {
  vpc_id = numspot_vpc.net.id
}`, netIpRange, subnetIpRange)
}

func testRouteTableConfig_Update(netIpRange, subnetIpRange string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = %[2]q
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.net.id
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.net.id
  subnet_id = numspot_subnet.subnet.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}`, netIpRange, subnetIpRange)
}

func TestAccRouteTableResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	netIpRange := "10.101.0.0/16"
	subnetIpRange := "10.101.1.0/24"
	tagKey := "name"
	tagValue := "Terraform-Test-Volume"
	tagValueUpdate := "Terraform-Test-Volume-Update"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testRouteTableConfig_Tags(netIpRange, subnetIpRange, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_route_table.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testRouteTableConfig_Tags(netIpRange, subnetIpRange, tagKey, tagValueUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.0.value", tagValueUpdate),
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testRouteTableConfig_Tags(netIpRange, subnetIpRange, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_route_table" "test" {
  vpc_id = numspot_vpc.net.id

  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, netIpRange, subnetIpRange, tagKey, tagValue)
}
