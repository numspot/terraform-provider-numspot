package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccDHCPOptionsDatasource_Basic(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	domainName := "foo.bar"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsDatasourceConfig_Basic(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
						"id":          acctest.PAIR_PREFIX + "numspot_dhcp_options.test.id",
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
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	domainName1 := "foo.bar"
	domainName2 := "null.local"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsDatasourceConfig_ByID(domainName1, domainName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "items.#", "2"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
						"id":          acctest.PAIR_PREFIX + "numspot_dhcp_options.obj1.id",
						"domain_name": domainName1,
					}),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
						"id":          acctest.PAIR_PREFIX + "numspot_dhcp_options.obj2.id",
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
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	domainName := "numspot.dev"
	tagName := "Name"
	tagValue := "dhcp_numspot"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsDatasourceConfig_WithTags(domainName, tagName, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
						"id":           acctest.PAIR_PREFIX + "numspot_dhcp_options.test.id",
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
