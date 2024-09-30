package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

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
			{ // 1 - Create testing
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
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"), // Check value for all resource attributes
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "vpc_to_virtual_gateway_links.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
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
				ResourceName:            "numspot_virtual_gateway.test",
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
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"), // Check value for all resource attributes
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "vpc_to_virtual_gateway_links.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
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
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "test_new" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test_new.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", "ipsec.1"), // Check value for all resource attributes
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "vpc_to_virtual_gateway_links.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", "numspot_vpc.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
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
