package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccSubnetResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	subnetDependencies := `
resource "numspot_vpc" "terraform-dep-vpc-subnet" {
  ip_range = "10.101.0.0/16"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Step 1 - Create subnet
			{
				Config: subnetDependencies + `
resource "numspot_subnet" "terraform-subnet-acctest" {
  vpc_id                  = numspot_vpc.terraform-dep-vpc-subnet.id
  ip_range                = "10.101.1.0/24"
  tags = [
{
      key   = "name"
      value = "terraform-subnet-acctest"
    }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_subnet.terraform-subnet-acctest", "vpc_id", "numspot_vpc.terraform-dep-vpc-subnet", "id"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "ip_range", "10.101.1.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.terraform-subnet-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-subnet-acctest",
					}),
				),
			},
			// Step 2 - Import subnet
			{
				ResourceName:            "numspot_subnet.terraform-subnet-acctest",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Step 3 - Update subnet attributes
			{
				Config: subnetDependencies + `
resource "numspot_subnet" "terraform-subnet-acctest" {
  vpc_id                  = numspot_vpc.terraform-dep-vpc-subnet.id
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = "true"
  tags = [{
      key   = "name"
      value = "terraform-subnet-acctest-update"
    }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_subnet.terraform-subnet-acctest", "vpc_id", "numspot_vpc.terraform-dep-vpc-subnet", "id"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "ip_range", "10.101.1.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.terraform-subnet-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-subnet-acctest-update",
					}),
				),
			},
			// Step 4 - Reset subnet
			{
				Config: ` `,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Step 5 - Create subnet with attributes
			{
				Config: subnetDependencies + `
resource "numspot_subnet" "terraform-subnet-acctest" {
  vpc_id                  = numspot_vpc.terraform-dep-vpc-subnet.id
  availability_zone_name =  "cloudgouv-eu-west-1a"
  ip_range                = "10.101.2.0/24"
  map_public_ip_on_launch = "true"
  tags = [{
      key   = "name"
      value = "terraform-subnet-acctest"
    }]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_subnet.terraform-subnet-acctest", "vpc_id", "numspot_vpc.terraform-dep-vpc-subnet", "id"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "ip_range", "10.101.2.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.terraform-subnet-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-subnet-acctest",
					}),
				),
			},
			// Step 6 - Replace subnet attributes
			{
				Config: subnetDependencies + `
resource "numspot_subnet" "terraform-subnet-acctest" {
  vpc_id                  = numspot_vpc.terraform-dep-vpc-subnet.id
  availability_zone_name =  "cloudgouv-eu-west-1b"
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = "true"
  tags = [{
      key   = "name"
      value = "terraform-subnet-acctest-replace"
    }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_subnet.terraform-subnet-acctest", "vpc_id", "numspot_vpc.terraform-dep-vpc-subnet", "id"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "availability_zone_name", "cloudgouv-eu-west-1b"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "ip_range", "10.101.1.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.terraform-subnet-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-subnet-acctest-replace",
					}),
				),
			},
			// Step 7 - Recreate subnet
			{
				Config: subnetDependencies + `
resource "numspot_subnet" "terraform-subnet-acctest-recreate" {
  vpc_id                  = numspot_vpc.terraform-dep-vpc-subnet.id
  availability_zone_name =  "cloudgouv-eu-west-1b"
  ip_range                = "10.101.2.0/24"
  map_public_ip_on_launch = "true"
  tags = [{
      key   = "name"
      value = "terraform-subnet-acctest-recreate"
    }]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_subnet.terraform-subnet-acctest-recreate", "vpc_id", "numspot_vpc.terraform-dep-vpc-subnet", "id"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest-recreate", "availability_zone_name", "cloudgouv-eu-west-1b"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest-recreate", "ip_range", "10.101.2.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest-recreate", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr("numspot_subnet.terraform-subnet-acctest-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.terraform-subnet-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-subnet-acctest-recreate",
					})),
			},
		},
	})
}
