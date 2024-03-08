package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccSubnetsDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchSubnetsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_subnets.testdata", "subnets.#", "1"),
				),
			},
		},
	})
}

func fetchSubnetsConfig() string {
	return `
data "numspot_subnets" "testdata" {
	vpc_ids = [numspot_vpc.main.id]
	depends_on 	= [numspot_vpc.main, numspot_subnet.test]
}
resource "numspot_vpc" "main" {
	ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
	vpc_id 		= numspot_vpc.main.id
	ip_range 	= "10.101.1.0/24"
}
`
}
