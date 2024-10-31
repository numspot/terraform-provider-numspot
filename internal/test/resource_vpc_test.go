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

	vpcReplaceDependcies := `
resource "numspot_dhcp_options" "terraform-dep-dhcp-options-vpc" {
  domain_name = "domain name"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Step 1 - Create VPC
			{
				Config: `
resource "numspot_vpc" "terraform-vpc-acctest" {
  ip_range = "10.101.0.0/16"
  tags = [{
    key   = "name"
    value = "terraform-vpc-acctest"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "ip_range", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.terraform-vpc-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-vpc-acctest",
					}),
				),
			},
			// Step 2 - Import
			{
				ResourceName:            "numspot_vpc.terraform-vpc-acctest",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Step 3 - Update VPC
			{
				Config: `
resource "numspot_vpc" "terraform-vpc-acctest" {
  ip_range = "10.101.0.0/16"
  tags = [{
    key   = "name"
    value = "terraform-vpc-acctest-update"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "ip_range", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.terraform-vpc-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-vpc-acctest-update",
					}),
				),
			},
			// Step 4 - Replace VPC
			{
				Config: vpcReplaceDependcies + `
resource "numspot_vpc" "terraform-vpc-acctest" {
  ip_range            = "10.102.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.terraform-dep-dhcp-options-vpc.id
  tenancy             = "dedicated"
  tags = [{
    key   = "name"
    value = "terraform-vpc-acctest-replace"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "ip_range", "10.102.0.0/16"),
					resource.TestCheckResourceAttrPair("numspot_vpc.terraform-vpc-acctest", "dhcp_options_set_id", "numspot_dhcp_options.terraform-dep-dhcp-options-vpc", "id"),
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "tenancy", "dedicated"),
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.terraform-vpc-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-vpc-acctest-replace",
					}),
				),
			},
			// Step 5 - Reset
			{
				Config: vpcReplaceDependcies + ` `,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Step 6 - Create with attributes
			{
				Config: vpcReplaceDependcies + `
resource "numspot_vpc" "terraform-vpc-acctest" {
  ip_range            = "10.101.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.terraform-dep-dhcp-options-vpc.id
  tenancy             = "dedicated"
  tags = [{
    key   = "name"
    value = "terraform-vpc-acctest"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "ip_range", "10.101.0.0/16"),
					resource.TestCheckResourceAttrPair("numspot_vpc.terraform-vpc-acctest", "dhcp_options_set_id", "numspot_dhcp_options.terraform-dep-dhcp-options-vpc", "id"),
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "tenancy", "dedicated"),
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.terraform-vpc-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-vpc-acctest",
					}),
				),
			},
			// Step 7 - Recreate VPC
			{
				Config: vpcReplaceDependcies + `
resource "numspot_vpc" "terraform-vpc-acctest-recreate" {
  ip_range            = "10.101.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.terraform-dep-dhcp-options-vpc.id
  tenancy             = "dedicated"
  tags = [{
    key   = "name"
    value = "terraform-vpc-acctest-recreate"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest-recreate", "ip_range", "10.101.0.0/16"),
					resource.TestCheckResourceAttrPair("numspot_vpc.terraform-vpc-acctest-recreate", "dhcp_options_set_id", "numspot_dhcp_options.terraform-dep-dhcp-options-vpc", "id"),
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest-recreate", "tenancy", "dedicated"),
					resource.TestCheckResourceAttr("numspot_vpc.terraform-vpc-acctest-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.terraform-vpc-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-vpc-acctest-recreate",
					})),
			},
		},
	})
}
