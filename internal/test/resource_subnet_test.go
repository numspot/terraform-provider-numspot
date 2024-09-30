package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
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

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create testing
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = "true"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Subnet"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr("numspot_subnet.test", "ip_range", "10.101.1.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Subnet",
					}),
					resource.TestCheckResourceAttrPair("numspot_subnet.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_subnet.test",
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

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = "false"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Subnet-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "map_public_ip_on_launch", "false"),
					resource.TestCheckResourceAttr("numspot_subnet.test", "ip_range", "10.101.1.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Subnet-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_subnet.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
			},
			// 4 - Update testing With Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  ip_range                = "10.101.2.0/24"
  map_public_ip_on_launch = "false"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Subnet-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "map_public_ip_on_launch", "false"),
					resource.TestCheckResourceAttr("numspot_subnet.test", "ip_range", "10.101.2.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Subnet-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_subnet.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						resourceId = v
						return nil
					})),
			},

			// <== If resource has required dependencies ==>
			{ // 5 - Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = "true"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Subnet"
    }
  ]
}`,
			},
			// 6 - Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "test_new" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test_new.id
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = "true"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Subnet"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.test", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr("numspot_subnet.test", "ip_range", "10.101.1.0/24"),
					resource.TestCheckResourceAttr("numspot_subnet.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Subnet",
					}),
					resource.TestCheckResourceAttrPair("numspot_subnet.test", "vpc_id", "numspot_vpc.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						return nil
					})),
			},
		},
	})
}
