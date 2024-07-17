//go:build acc

package provider

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
)

func TestAccDHCPOptionsDatasource_Basic(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	domainName := fmt.Sprintf("foo.bar.%s", strconv.FormatInt(time.Now().UnixMilli(), 10))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsDatasourceConfig_Basic(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "items.#", "1"),
					utils_acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
						"id":          utils_acctest.PAIR_PREFIX + "numspot_dhcp_options.test.id",
						"domain_name": domainName,
					}),
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

data "numspot_dhcp_options" "testdata" {
  domain_names = [numspot_dhcp_options.test.domain_name]
}`, domainName)
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
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "items.#", "2"),
					utils_acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
						"id":          utils_acctest.PAIR_PREFIX + "numspot_dhcp_options.obj1.id",
						"domain_name": domainName1,
					}),
					utils_acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
						"id":          utils_acctest.PAIR_PREFIX + "numspot_dhcp_options.obj2.id",
						"domain_name": domainName2,
					}),
				),
			},
		},
	})
}

func testAccDHCPOptionsDatasourceConfig_ByID(domainName1, domainName2 string) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "obj1" {
  domain_name = %[1]q
}
resource "numspot_dhcp_options" "obj2" {
  domain_name = %[2]q
}

data "numspot_dhcp_options" "testdata" {
  ids = [numspot_dhcp_options.obj1.id, numspot_dhcp_options.obj2.id]
}`, domainName1, domainName2)
}

func TestAccDHCPOptionsDatasource_WithTags(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	domainName := "numspot.dev"
	tagName := fmt.Sprintf("Name-%s", strconv.FormatInt(time.Now().UnixMilli(), 10))
	tagValue := "dhcp_numspot"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsDatasourceConfig_WithTags(domainName, tagName, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "items.#", "1"),
					utils_acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
						"id":           utils_acctest.PAIR_PREFIX + "numspot_dhcp_options.test.id",
						"domain_name":  domainName,
						"tags.0.key":   tagName,
						"tags.0.value": tagValue,
					}),
				),
			},
		},
	})
}

func testAccDHCPOptionsDatasourceConfig_WithTags(domainName, tagName, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "test" {
  domain_name = %[1]q
  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}
data "numspot_dhcp_options" "testdata" {
  tags = [
    format("%%s=%%s", numspot_dhcp_options.test.tags[0].key, numspot_dhcp_options.test.tags[0].value)
  ]
}`, domainName, tagName, tagValue)
}
