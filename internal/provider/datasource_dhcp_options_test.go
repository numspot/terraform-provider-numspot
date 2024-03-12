package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDHCPOptionsDatasource_Basic(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	domainName := "foo.bar"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsDatasourceConfig_Basic(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.0.domain_name", domainName),
				),
			},
		},
	})
}

func TestAccDHCPOptionsDatasource_ByID(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	domainName1 := "foo.bar"
	domainName2 := "null.local"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsDatasourceConfig_ByID(domainName1, domainName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.#", "2"),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.0.domain_name", domainName1),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.1.domain_name", domainName2),
				),
			},
		},
	})
}

func TestAccDHCPOptionsDatasource_WithTags(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	domainName := "numspot.dev"
	tagName := "Name"
	tagValue := "dhcp_numspot"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsDatasourceConfig_WithTags(domainName, tagName, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.0.domain_name", domainName),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.0.tags.0.key", tagName),
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.test", "dhcp_options.0.tags.0.value", tagValue),
				),
			},
		},
	})
}

func testAccDHCPOptionsDatasourceConfig_Basic(domainName string) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "test" {
		domain_name = %[1]q
	}

data "numspot_dhcp_options" "test" {
	domain_names = [numspot_dhcp_options.test.domain_name]
}`, domainName)
}

func testAccDHCPOptionsDatasourceConfig_ByID(domainName1, domainName2 string) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "obj1" {
		domain_name = %[1]q
}
resource "numspot_dhcp_options" "obj2" {
		domain_name = %[2]q
}

data "numspot_dhcp_options" "test" {
	ids = [numspot_dhcp_options.obj1.id, numspot_dhcp_options.obj2.id]
}`, domainName1, domainName2)
}

func testAccDHCPOptionsDatasourceConfig_WithTags(domainName, tagName, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "test" {
	domain_name = %[1]q
	tags = [
		{
			key = %[2]q
			value = %[3]q
		}
	]
}
data "numspot_dhcp_options" "test" {
	tags = [
format("%%s=%%s", numspot_dhcp_options.test.tags[0].key, numspot_dhcp_options.test.tags[0].value)
]
}`, domainName, tagName, tagValue)
}
