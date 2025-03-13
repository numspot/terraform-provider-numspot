package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// 1 - Create unlinked virtual gateway
// 2 - import
// 3 - Update unlinked virtual gateway (tags)
// 4 - Recreate unlinked virtual gateway
//
// 5 - Link unlinked virtual gateway to a vpc
// 6 - Update attributes from linked virtual gateway
// 7 - Recreate virtual gateway
// 8 - Unlink linked virtual gateway
//
// 9 - setup
// 10 - Unlink and link virtual gateway to a new VPC
// 11 - Unlink and link virtual gateway to a new VPC with deletion of old one
func TestAccVirtualGatewayResource(t *testing.T) {
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
			{ // 1 - Create unlinked virtual gateway
				Config: `
resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			{ // 2 - ImportState testing
				ResourceName:            "numspot_virtual_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{ // 3 - Update unlinked virtual gateway (tags)
				Config: `
resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 4 - Recreate unlinked virtual gateway
				Config: `
resource "numspot_virtual_gateway" "test_recreated" {
  connection_type = "ipsec.1"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test_recreated", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test_recreated", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 5 - Link unlinked virtual gateway to a vpc
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "vpc_to_virtual_gateway_links.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 6 - Update attributes from linked virtual gateway
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "vpc_to_virtual_gateway_links.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 7 - Recreate virtual gateway
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test_recreated" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test_recreated", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test_recreated", "vpc_to_virtual_gateway_links.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test_recreated", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test_recreated", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 8 - Unlink linked virtual gateway
				Config: `
resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "vpc_to_virtual_gateway_links.#", "0"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 9 - setup for next step
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 10 - Unlink and link virtual gateway to a new VPC
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}
resource "numspot_vpc" "test_new" {
  ip_range = "10.102.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test_new.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "vpc_to_virtual_gateway_links.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", "numspot_vpc.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 11 - Unlink and link virtual gateway to a new VPC with deletion of old one
				Config: `
resource "numspot_vpc" "test_newest" {
  ip_range = "10.104.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test_newest.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "vpc_to_virtual_gateway_links.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", "numspot_vpc.test_newest", "id"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
