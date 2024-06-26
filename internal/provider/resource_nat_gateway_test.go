//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNatGatewayResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNatGatewayConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("numspot_net_access_point.test", "field", "value"),
				//resource.TestCheckResourceAttrWith("numspot_net_access_point.test", "field", func(v string) error {
				//	require.NotEmpty(t, v)
				//	return nil
				//}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nat_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testNatGatewayConfig() string {
	return `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test" {}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test.id
  public_ip_id = numspot_public_ip.test.id
  depends_on   = [numspot_route_table.test]
}
`
}

func TestAccNatGatewayResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	tagKey := "name"
	tagValue := "Terraform-Test-Nat-Gateway"
	tagValueUpdated := tagValue + "-Updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Create testing
			{
				Config: testNatGatewayConfig_Tags(tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nat_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNatGatewayConfig_Tags(tagKey, tagValueUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.0.value", tagValueUpdated),
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testNatGatewayConfig_Tags(key, value string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test" {}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test.id
  public_ip_id = numspot_public_ip.test.id

  tags = [
    {
      key   = %[1]q
      value = %[2]q
    }
  ]
}`, key, value)
}
