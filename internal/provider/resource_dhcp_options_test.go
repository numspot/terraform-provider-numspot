package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDhcpOptionsResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	domainName := "foo.bar"
	updatedDomainName := "bar.foo"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testDhcpOptionsConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "domain_name", domainName),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_dhcp_options.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testDhcpOptionsConfig(updatedDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "domain_name", updatedDomainName),
				),
			},
		},
	})
}

func testDhcpOptionsConfig(domainName string) string {
	return fmt.Sprintf(`resource "numspot_dhcp_options" "test" {
	domain_name = %[1]q
}`, domainName)
}
