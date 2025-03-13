package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test steps:
// 1 - Create testing
// 2 - Update testing Without Replace
// 3 - Update testing With Replace
// 4 - recreate testing
// 5 - reset: unset tags
// 6 - set tags
func TestAccVpcPeeringResource(t *testing.T) {
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
resource "numspot_vpc" "accepter" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "source" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter.id
  source_vpc_id   = numspot_vpc.source.id

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Vpc-Peering"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc_peering.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Vpc-Peering",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test", "accepter_vpc_id", "numspot_vpc.accepter", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test", "source_vpc_id", "numspot_vpc.source", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			// 2 - Update testing Without Replace
			{
				Config: `
resource "numspot_vpc" "accepter" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "source" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter.id
  source_vpc_id   = numspot_vpc.source.id

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Vpc-Peering-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc_peering.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Vpc-Peering-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test", "accepter_vpc_id", "numspot_vpc.accepter", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test", "source_vpc_id", "numspot_vpc.source", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},

			// <== If resource has required dependencies ==>
			// 3 - Update testing With Replace of dependency resource and with Replace of the resource
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "accepter_new" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "source_new" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter_new.id
  source_vpc_id   = numspot_vpc.source_new.id

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Vpc-Peering"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc_peering.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Vpc-Peering",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test", "accepter_vpc_id", "numspot_vpc.accepter_new", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test", "source_vpc_id", "numspot_vpc.source_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			// 4 - recreate testing
			{
				Config: `
resource "numspot_vpc" "accepter_new" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "source_new" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test-recreate" {
  accepter_vpc_id = numspot_vpc.accepter_new.id
  source_vpc_id   = numspot_vpc.source_new.id

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Vpc-Peering"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc_peering.test-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc_peering.test-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Vpc-Peering",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test-recreate", "accepter_vpc_id", "numspot_vpc.accepter_new", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test-recreate", "source_vpc_id", "numspot_vpc.source_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc_peering.test-recreate", "id", func(v string) error {
						err := acctest.CheckResourceIdChanged(t, v, &resourceId)
						resourceId = v
						return err
					}),
				),
			},
			// 5 - reset: unset tags
			{
				Config: `
resource "numspot_vpc" "accepter_new" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "source_new" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test-recreate" {
  accepter_vpc_id = numspot_vpc.accepter_new.id
  source_vpc_id   = numspot_vpc.source_new.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc_peering.test-recreate", "tags.#", "0"),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test-recreate", "accepter_vpc_id", "numspot_vpc.accepter_new", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test-recreate", "source_vpc_id", "numspot_vpc.source_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc_peering.test-recreate", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// 6 - set tags
			{
				Config: `
resource "numspot_vpc" "accepter_new" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "source_new" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test-recreate" {
  accepter_vpc_id = numspot_vpc.accepter_new.id
  source_vpc_id   = numspot_vpc.source_new.id

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Vpc-Peering"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc_peering.test-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc_peering.test-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Vpc-Peering",
					}),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test-recreate", "accepter_vpc_id", "numspot_vpc.accepter_new", "id"),
					resource.TestCheckResourceAttrPair("numspot_vpc_peering.test-recreate", "source_vpc_id", "numspot_vpc.source_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_vpc_peering.test-recreate", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
