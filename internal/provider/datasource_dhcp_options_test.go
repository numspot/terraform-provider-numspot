package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDHCPOptionsDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchDHCPOptionsByDomainNamesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "dhcp_options.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "dhcp_options.0.domain_name", "foo.bar"),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdatabyid", "dhcp_options.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdatabyid", "dhcp_options.0.domain_name", "foo.bar"),
				),
			},
		},
	})
}

func fetchDHCPOptionsByDomainNamesConfig() string {
	return `
resource "numspot_dhcp_options" "test" {
		domain_name = "foo.bar"
	}

data "numspot_dhcp_options" "testdata" {
	domain_names = [numspot_dhcp_options.test.domain_name]
}

data "numspot_dhcp_options" "testdatabyid" {
	ids = [numspot_dhcp_options.test.id]
}`

}
