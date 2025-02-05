package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test cases:
// 1 - Create route table with routes
// 2 - ImportState testing
// 3 - Update testing Without Replace (if needed)
// 4 - Update testing With Replace of dependency resource and with Replace of the resource
// 5 - recreate testing
// 6 - reset
// 7 - Test update: add routes and tags

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
			{ // 1 - Create route table with routes
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
						"gateway_id":           acctest.PairPrefix + "numspot_internet_gateway.test.id",
						"destination_ip_range": "0.0.0.0/0",
					}),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
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
      destination_ip_range = "10.0.0.0/16"
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
						"gateway_id":           acctest.PairPrefix + "numspot_internet_gateway.test.id",
						"destination_ip_range": "10.0.0.0/16",
					}),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},

			// 4 - Update testing With Replace of dependency resource and with Replace of the resource
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
						"gateway_id":           acctest.PairPrefix + "numspot_internet_gateway.test_new.id",
						"destination_ip_range": "0.0.0.0/0",
					}),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			// 5 - recreate testing
			{
				Config: `
resource "numspot_vpc" "test_recreate" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test_recreate" {
  vpc_id   = numspot_vpc.test_recreate.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_internet_gateway" "test_recreate" {
  vpc_id = numspot_vpc.test_recreate.id
}

resource "numspot_route_table" "test_recreate" {
  vpc_id    = numspot_vpc.test_recreate.id
  subnet_id = numspot_subnet.test_recreate.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test_recreate.id
    }
  ]
  tags = [{
    key   = "name"
    value = "Terraform-Test-Volume-recreated"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-recreated",
					}),
					resource.TestCheckResourceAttrPair("numspot_route_table.test_recreate", "subnet_id", "numspot_subnet.test_recreate", "id"),
					resource.TestCheckResourceAttrPair("numspot_route_table.test_recreate", "vpc_id", "numspot_vpc.test_recreate", "id"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_route_table.test_recreate", "routes.*", map[string]string{
						"gateway_id":           acctest.PairPrefix + "numspot_internet_gateway.test_recreate.id",
						"destination_ip_range": "0.0.0.0/0",
					}),
					resource.TestCheckResourceAttrWith("numspot_route_table.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			// 6 - reset
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_route_table" "test" {
  vpc_id = numspot_vpc.test.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_route_table.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			// 7 - Test update: add routes and tags
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
						"gateway_id":           acctest.PairPrefix + "numspot_internet_gateway.test.id",
						"destination_ip_range": "0.0.0.0/0",
					}),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
