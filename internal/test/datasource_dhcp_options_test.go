package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccDHCPOptionsDataSource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Step 1 - Create DHCP Options with domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 2 - Update DHCP Options with domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 3 - Create DHCP Options with domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 4 - Create DHCP Options with domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 5 - Create DHCP Options with domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 6 - Create DHCP Options with domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 7 - Create DHCP Options with domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
		},
	})
}

//func TestAccDHCPOptionsDatasource_ByID(t *testing.T) {
//	acct := acctest.NewAccTest(t, false, "")
//	defer func() {
//		err := acct.Cleanup()
//		require.NoError(t, err)
//	}()
//	pr := acct.TestProvider
//
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: pr,
//		Steps: []resource.TestStep{
//			{
//				Config: `
//resource "numspot_dhcp_options" "obj1" {
//  domain_name = "foo.bar"
//}
//resource "numspot_dhcp_options" "obj2" {
//  domain_name = "null.local"
//}
//
//data "numspot_dhcp_options" "testdata" {
//  ids = [numspot_dhcp_options.obj1.id, numspot_dhcp_options.obj2.id]
//}`,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "items.#", "2"),
//					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
//						"id":          acctest.PAIR_PREFIX + "numspot_dhcp_options.obj1.id",
//						"domain_name": "foo.bar",
//					}),
//					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
//						"id":          acctest.PAIR_PREFIX + "numspot_dhcp_options.obj2.id",
//						"domain_name": "null.local",
//					}),
//				),
//			},
//		},
//	})
//}
//
//func TestAccDHCPOptionsDatasource_WithTags(t *testing.T) {
//	acct := acctest.NewAccTest(t, false, "")
//	defer func() {
//		err := acct.Cleanup()
//		require.NoError(t, err)
//	}()
//	pr := acct.TestProvider
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: pr,
//		Steps: []resource.TestStep{
//			{
//				Config: `
//resource "numspot_dhcp_options" "test" {
//  domain_name = "numspot.dev"
//  tags = [
//    {
//      key   = "Name"
//      value = "dhcp_numspot"
//    }
//  ]
//}
//data "numspot_dhcp_options" "testdata" {
//  tags = [
//    "Name=dhcp_numspot"
//  ]
//  depends_on = [numspot_dhcp_options.test]
//}`,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdata", "items.#", "1"),
//					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_dhcp_options.testdata", "items.*", map[string]string{
//						"id":           acctest.PAIR_PREFIX + "numspot_dhcp_options.test.id",
//						"domain_name":  "numspot.dev",
//						"tags.0.key":   "Name",
//						"tags.0.value": "dhcp_numspot",
//					}),
//				),
//			},
//		},
//	})
//}
