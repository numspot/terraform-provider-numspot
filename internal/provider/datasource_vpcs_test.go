package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVPCsDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVPCsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpcs.testdata", "vpcs.#", "1"),
				),
			},
		},
	})
}

func fetchVPCsConfig() string {
	return `
resource "numspot_vpc" "test" {
	ip_range = "10.101.0.0/16"
}

data "numspot_vpcs" "testdata" {
	ids = [numspot_vpc.test.id]
}`
}
