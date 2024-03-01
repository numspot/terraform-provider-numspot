package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetPeeringResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetPeeringConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("numspot_vpc_peering.test", "field", "value"),
				//resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "field", func(v string) error {
				//	require.NotEmpty(t, v)
				//	return nil
				//}),
				),
			},
			// ImportState testing
			// Update testing
			//{
			//	Config: testNetPeeringConfig(),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("numspot_vpc_peering.test", "field", "value"),
			//		resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "field", func(v string) error {
			//			return nil
			//		}),
			//	),
			//},
		},
	})
}

func testNetPeeringConfig() string {
	return `
resource "numspot_vpc" "source" {
	ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "accepter" {
	ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
	accepter_vpc_id 	= numspot_vpc.accepter.id
	source_vpc_id 		= numspot_vpc.source.id
}`
}
