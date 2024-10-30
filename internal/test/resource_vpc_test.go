package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVpcResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create testing
				Config: `
resource "numspot_dhcp_options" "test" {
  domain_name = "domain"
}

resource "numspot_vpc" "test" {
  ip_range            = "10.101.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.test.id
  tenancy             = "default"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tenancy", "default"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc.test", "dhcp_options_set_id", "numspot_dhcp_options.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_vpc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace
			{
				Config: `
resource "numspot_dhcp_options" "test" {
  domain_name = "domain"
}

resource "numspot_vpc" "test" {
  ip_range            = "10.101.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.test.id
  tenancy             = "default"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tenancy", "default"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc.test", "dhcp_options_set_id", "numspot_dhcp_options.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// 4 - Update testing With Replace
			{
				Config: `
resource "numspot_dhcp_options" "test" {
  domain_name = "domain"
}

resource "numspot_vpc" "test" {
  ip_range            = "10.102.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.test.id
  tenancy             = "dedicated"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", "10.102.0.0/16"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tenancy", "dedicated"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc.test", "dhcp_options_set_id", "numspot_dhcp_options.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},

			// <== If resource has required dependencies ==>
			// 5 - Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_dhcp_options" "test_new" {
  domain_name = "domain"
}

resource "numspot_vpc" "test" {
  ip_range            = "10.102.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.test_new.id
  tenancy             = "dedicated"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", "10.102.0.0/16"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tenancy", "dedicated"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc.test", "dhcp_options_set_id", "numspot_dhcp_options.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
