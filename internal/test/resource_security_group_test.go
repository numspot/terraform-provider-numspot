//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccSecurityGroupResource(t *testing.T) {
	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// 1 - Create Security group
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test.id
  name        = "security-group-name"
  description = "security-group-description"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			// 2 - Import
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test.id
  name        = "security-group-name"
  description = "security-group-description"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// 4 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test.id
  name        = "security-group-name"
  description = "security-group-description"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// 5 - Update testing With Replace
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test.id
  name        = "security-group-name-updated"
  description = "security-group-description-updated"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name-updated"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description-updated"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			// <== If resource has required dependencies ==>
			// 6 - Update testing With Replace of dependency resource and with Replace of the resource
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "test_new" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test_new.id
  name        = "security-group-name-updated"
  description = "security-group-description-updated"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name-updated"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description-updated"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			// 7- recreate testing
			{
				Config: `
resource "numspot_vpc" "test_recreate" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test_recreate" {
  vpc_id      = numspot_vpc.test_recreate.id
  name        = "security-group-name"
  description = "security-group"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "description", "security-group"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test_recreate", "vpc_id", "numspot_vpc.test_recreate", "id"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			// 8- reset rules
			{
				Config: `
resource "numspot_vpc" "test_recreate" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test_recreate" {
  vpc_id      = numspot_vpc.test_recreate.id
  name        = "security-group-name"
  description = "security-group"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "description", "security-group"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test_recreate", "vpc_id", "numspot_vpc.test_recreate", "id"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// 9- Add rules
			{
				Config: `
resource "numspot_vpc" "test_recreate" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test_recreate" {
  vpc_id      = numspot_vpc.test_recreate.id
  name        = "security-group-name"
  description = "security-group"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "description", "security-group"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test_recreate", "vpc_id", "numspot_vpc.test_recreate", "id"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
