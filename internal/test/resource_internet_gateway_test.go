package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccInternetGatewayResource(t *testing.T) {
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
			// 1 - Create testing
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-internetgateway-acctest"
    }
  ]
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_internet_gateway.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttr("numspot_internet_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.test", "id", func(v string) error {
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
				ResourceName:            "numspot_internet_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-internetgateway-acctest"
    }
  ]
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_internet_gateway.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttr("numspot_internet_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.test", "id", func(v string) error {
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
			// 4 - Update testing With Replace of VPC and without Replacing Internet Gateway (if needed)
			// This test is useful to check wether or not the deletion of the VPC and then the update of the Internet Gateway works properly
			{
				Config: `
resource "numspot_vpc" "test_new" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-internetgateway-acctest"
    }
  ]
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test_new.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_internet_gateway.test", "vpc_id", "numspot_vpc.test_new", "id"),
					resource.TestCheckResourceAttr("numspot_internet_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.test", "id", func(v string) error {
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
