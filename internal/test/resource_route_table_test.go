package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccRouteTableResource(t *testing.T) {
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
  vpc_id   = numspot_vpc.test.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
  tags = [{
    key   = "name"
    value = "Terraform-Test-Volume"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttrPair("numspot_route_table.test", "subnet_id", "numspot_subnet.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_route_table.test", "vpc_id", "numspot_vpc.test", "id"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_route_table.test", "routes.*", map[string]string{
						"gateway_id": acctest.PAIR_PREFIX + "numspot_internet_gateway.test.id",
					}),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
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
				ResourceName:            "numspot_route_table.test",
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
  vpc_id   = numspot_vpc.test.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
  tags = [{
    key   = "name"
    value = "Terraform-Test-Volume-Updated"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_route_table.test", "subnet_id", "numspot_subnet.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_route_table.test", "vpc_id", "numspot_vpc.test", "id"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_route_table.test", "routes.*", map[string]string{
						"gateway_id": acctest.PAIR_PREFIX + "numspot_internet_gateway.test.id",
					}),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
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

			// <== If resource has optional dependencies ==>
			// 4 - Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
			{
				Config: `
resource "numspot_vpc" "test_new" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test_new" {
  vpc_id   = numspot_vpc.test_new.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_internet_gateway" "test_new" {
  vpc_id = numspot_vpc.test_new.id
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test_new.id
  subnet_id = numspot_subnet.test_new.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test_new.id
    }
  ]
  tags = [{
    key   = "name"
    value = "Terraform-Test-Volume-Updated"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_route_table.test", "subnet_id", "numspot_subnet.test_new", "id"),
					resource.TestCheckResourceAttrPair("numspot_route_table.test", "vpc_id", "numspot_vpc.test_new", "id"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_route_table.test", "routes.*", map[string]string{
						"gateway_id": acctest.PAIR_PREFIX + "numspot_internet_gateway.test_new.id",
					}),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						return nil
					}),
				),
			},
		},
	})
}
