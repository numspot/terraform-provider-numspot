package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNatGatewayResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNatGatewayConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_net_access_point.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_net_access_point.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNatGatewayConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_net_access_point.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testNatGatewayConfig() string {
	return `
resource "numspot_net" "test" {
	ip_range = "10.101.0.0/16"
}

resource "numspot_internet_service" "test" {
	net_id=numspot_net.test.id
}

resource "numspot_subnet" "test" {
	net_id = numspot_net.test.id
	map_public_ip_on_launch = true
	ip_range = "10.101.1.0/24"
}

resource "numspot_public_ip" "test" {}

resource "numspot_route_table" "test" {
	net_id =  numspot_net.test.id
	subnet_id = numspot_subnet.test.id
	routes = [
		{
			destination_ip_range 	= "0.0.0.0/0"
			gateway_id 	 			= numspot_internet_service.test.id
		}
	]
}

resource "numspot_nat_gateway" "test" {
	subnet_id = numspot_subnet.test.id
	public_ip_id = numspot_public_ip.test.id

	depends_on = [numspot_route_table.test]
}
`
}
