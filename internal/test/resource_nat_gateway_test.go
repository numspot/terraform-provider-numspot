package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// 1 - Create NatGateway
// 2 - Import
// 3 - Update attributes
// 4 - Recreate NatGateway
// 5 - Associate NatGateway to a new publicIp and new Subnet (without deleting old ones)
// 6 - Associate NatGateway to a new publicIp and new Subnet (with delete of old ones)
// 7 - Recreate route table on VPC associated to the natgateway
func TestAccNatGatewayResource(t *testing.T) {
	acct := acctest.NewAccTest(t, true, "record")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create NatGateway
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test" {}
resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test.id
  public_ip_id = numspot_public_ip.test.id
  depends_on   = [numspot_route_table.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Nat-Gateway"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nat_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Nat-Gateway",
					}),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "subnet_id", "numspot_subnet.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "public_ip_id", "numspot_public_ip.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_nat_gateway.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			{ // 2 - ImportState testing
				ResourceName:            "numspot_nat_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			{ // 3 - Update attributes Without Replace
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test" {}
resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test.id
  public_ip_id = numspot_public_ip.test.id
  depends_on   = [numspot_route_table.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Nat-Gateway-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nat_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Nat-Gateway-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "subnet_id", "numspot_subnet.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "public_ip_id", "numspot_public_ip.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_nat_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 4 - Recreate NatGateway
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test" {}
resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

resource "numspot_nat_gateway" "test_recreated" {
  subnet_id    = numspot_subnet.test.id
  public_ip_id = numspot_public_ip.test.id
  depends_on   = [numspot_route_table.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Nat-Gateway"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_gateway.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nat_gateway.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Nat-Gateway",
					}),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test_recreated", "subnet_id", "numspot_subnet.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test_recreated", "public_ip_id", "numspot_public_ip.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_nat_gateway.test_recreated", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 5 - Associate NatGateway to a new publicIp and new Subnet (without deleting old ones) (note: the route_table will get replaced here)
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}


resource "numspot_subnet" "test_new" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.50.0/24"
}

resource "numspot_public_ip" "test" {}
resource "numspot_public_ip" "test_new" {}
resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test_new.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}
resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test_new.id
  public_ip_id = numspot_public_ip.test_new.id
  depends_on   = [numspot_route_table.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Nat-Gateway"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nat_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Nat-Gateway",
					}),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "subnet_id", "numspot_subnet.test_new", "id"),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "public_ip_id", "numspot_public_ip.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_nat_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 6 - Associate NatGateway to a new publicIp and new Subnet (with delete of old ones)
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test_newest" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test_newest" {}
resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test_newest.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}
resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test_newest.id
  public_ip_id = numspot_public_ip.test_newest.id
  depends_on   = [numspot_route_table.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Nat-Gateway"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nat_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Nat-Gateway",
					}),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "subnet_id", "numspot_subnet.test_newest", "id"),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "public_ip_id", "numspot_public_ip.test_newest", "id"),
					resource.TestCheckResourceAttrWith("numspot_nat_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 7 - Recreate route table on VPC associated to the natgateway without updating natgateway
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test_newest" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test_newest" {}
resource "numspot_route_table" "test_recreated" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test_newest.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}
resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test_newest.id
  public_ip_id = numspot_public_ip.test_newest.id
  depends_on   = [numspot_route_table.test_recreated]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Nat-Gateway"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nat_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Nat-Gateway",
					}),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "subnet_id", "numspot_subnet.test_newest", "id"),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "public_ip_id", "numspot_public_ip.test_newest", "id"),
					resource.TestCheckResourceAttrWith("numspot_nat_gateway.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
