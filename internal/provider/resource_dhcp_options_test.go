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
	tagName := "Terraform Provider DHCP Options"
	updatedTagName := "Terraform Provider DHCP Options - 2"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testDhcpOptionsConfig(domainName, tagName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "domain_name", domainName),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.0.key", "Name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.0.value", tagName),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.#", "1"),
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
				Config: testDhcpOptionsConfig(updatedDomainName, updatedTagName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "domain_name", updatedDomainName),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.0.key", "Name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.0.value", updatedTagName),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testDhcpOptionsConfig(domainName, tagValue string) string {
	return fmt.Sprintf(`resource "numspot_dhcp_options" "test" {
  domain_name = %[1]q
  tags = [
    {
      key   = "Name"
      value = %[2]q
    }
  ]
}`, domainName, tagValue)
}
