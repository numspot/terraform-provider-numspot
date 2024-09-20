package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccNatGatewayResource(t *testing.T) {
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
				ResourceName:            "numspot_nat_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace
			{
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

			// <== If resource has required dependencies ==>
			// Update testing With Replace of Subnet and PublicIP and with Replace of Nat Gateway
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test_new" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

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
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "subnet_id", "numspot_subnet.test_new", "id"),
					resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "public_ip_id", "numspot_public_ip.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_nat_gateway.test", "id", func(v string) error {
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
