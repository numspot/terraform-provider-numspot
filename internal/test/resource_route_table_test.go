//go:build acc

package test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

// Test cases:
// 1 - Create route table with routes
// 2 - ImportState testing
// 3 - Update testing Without Replace (if needed)
// 4 - Update testing With Replace of dependency resource and with Replace of the resource
// 5 - recreate testing
// 6 - reset
// 7 - Test update: add routes and tags
// 8 - Test subnet_ids with multiple subnets
// 9 - Test update subnet_ids (diff-based)
// 10 - Test empty subnet_ids
// 11 - Test mutual exclusivity error (subnet_id and subnet_ids together)

func TestAccRouteTableResource(t *testing.T) {
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
			{ // 1 - Create route table with routes
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
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
    value = "Terraform-Test-RouteTable"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-RouteTable",
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
				ImportStateVerifyIgnore: []string{"id", "subnet_id"},
			},
			// 3 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
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
    value = "Terraform-Test-RouteTable-Updated"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-RouteTable-Updated",
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
  vpc_id                 = numspot_vpc.test_new.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
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
    value = "Terraform-Test-RouteTable-Updated"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-RouteTable-Updated",
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
  vpc_id                 = numspot_vpc.test_recreate.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
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
    value = "Terraform-Test-RouteTable-recreated"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-RouteTable-recreated",
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
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
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
    value = "Terraform-Test-RouteTable"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-RouteTable",
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
			// 8 - Test subnet_ids with multiple subnets
			{
				Config: `
resource "numspot_vpc" "test_multi" {
  ip_range = "10.102.0.0/16"
}

resource "numspot_subnet" "test1" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.1.0/24"
  availability_zone_name = "eu-west-2a"
}

resource "numspot_subnet" "test2" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.2.0/24"
  availability_zone_name = "eu-west-2b"
}

resource "numspot_subnet" "test3" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.3.0/24"
  availability_zone_name = "eu-west-2a"
}

resource "numspot_route_table" "test_multi" {
  vpc_id     = numspot_vpc.test_multi.id
  subnet_ids = [numspot_subnet.test1.id, numspot_subnet.test2.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test_multi", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttrWith("numspot_route_table.test_multi", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			// 9 - Test update subnet_ids (diff-based)
			{
				Config: `
resource "numspot_vpc" "test_multi" {
  ip_range = "10.102.0.0/16"
}

resource "numspot_subnet" "test1" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.1.0/24"
  availability_zone_name = "eu-west-2a"
}

resource "numspot_subnet" "test2" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.2.0/24"
  availability_zone_name = "eu-west-2b"
}

resource "numspot_subnet" "test3" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.3.0/24"
  availability_zone_name = "eu-west-2a"
}

resource "numspot_route_table" "test_multi" {
  vpc_id     = numspot_vpc.test_multi.id
  subnet_ids = [numspot_subnet.test1.id, numspot_subnet.test3.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test_multi", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttrWith("numspot_route_table.test_multi", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// 10 - Test empty subnet_ids
			{
				Config: `
resource "numspot_vpc" "test_multi" {
  ip_range = "10.102.0.0/16"
}

resource "numspot_subnet" "test1" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.1.0/24"
  availability_zone_name = "eu-west-2a"
}

resource "numspot_subnet" "test2" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.2.0/24"
  availability_zone_name = "eu-west-2b"
}

resource "numspot_subnet" "test3" {
  vpc_id                 = numspot_vpc.test_multi.id
  ip_range               = "10.102.3.0/24"
  availability_zone_name = "eu-west-2a"
}

resource "numspot_route_table" "test_multi" {
  vpc_id     = numspot_vpc.test_multi.id
  subnet_ids = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test_multi", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrWith("numspot_route_table.test_multi", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// 11 - Test mutual exclusivity error (subnet_id and subnet_ids together)
			{
				Config: `
resource "numspot_vpc" "test_excl" {
  ip_range = "10.103.0.0/16"
}

resource "numspot_subnet" "test_excl" {
  vpc_id                 = numspot_vpc.test_excl.id
  ip_range               = "10.103.1.0/24"
  availability_zone_name = "eu-west-2a"
}

resource "numspot_route_table" "test_excl" {
  vpc_id     = numspot_vpc.test_excl.id
  subnet_id  = numspot_subnet.test_excl.id
  subnet_ids = [numspot_subnet.test_excl.id]
}`,
				ExpectError: regexp.MustCompile(`Attribute "subnet_ids" cannot be specified when "subnet_id" is specified`),
			},
		},
	})
}
