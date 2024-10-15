package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// / Test with unlinked internet gateway
// 1 - Create unlinked internet gateway
// 2 - importstate
// 3 - Update attributes from unlinked internet gateway
// 4 - Recreate unlinked internet gateway
//
// / Test with linked internet gateway
// 5 - Replace internet gateway by linking a vpc
// 6 - Update attributes from linked internet gateway
// 7 - Recreate linked internet gateway
// 8 - Unlink linked internet gateway
//
// / Other interaction tests with side resources
// 9 - Unlink and link internet gateway to a new vpc
// 10- Delete vpc and link internet gateway to a new vpc

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
			// 1 - Create unlinked internet gateway
			{
				Config: `
resource "numspot_internet_gateway" "test" {
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
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
				ImportStateVerifyIgnore: []string{},
			},
			// 3 - Update attributes from unlinked internet gateway
			{
				Config: `
resource "numspot_internet_gateway" "test" {
  tags = [
    {
      key   = "name-updated"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_internet_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test", "tags.*", map[string]string{
						"key":   "name-updated",
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
			// 4 - Recreate unlinked internet gateway
			{
				Config: `
resource "numspot_internet_gateway" "test_recreated" {
  tags = [
    {
      key   = "name-updated"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_internet_gateway.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test_recreated", "tags.*", map[string]string{
						"key":   "name-updated",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.test_recreated", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed.")
						}
						resourceId = v

						return nil
					}),
				),
			},
			// 5 - Replace internet gateway by linking a vpc
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
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed.")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 6 - Update attributes from linked internet gateway
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
      value = "Terraform-Test-Volume-Updated-Again"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_internet_gateway.test", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttr("numspot_internet_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated-Again",
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
			// 7 - Recreate linked internet gateway
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

resource "numspot_internet_gateway" "test_recreated" {
  vpc_id = numspot_vpc.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated-Again"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_internet_gateway.test_recreated", "vpc_id", "numspot_vpc.test", "id"),
					resource.TestCheckResourceAttr("numspot_internet_gateway.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated-Again",
					}),
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.test_recreated", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed.")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 8 - Unlink linked internet gateway
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

resource "numspot_internet_gateway" "test_recreated" {
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated-Again"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_internet_gateway.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated-Again",
					}),
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.test_recreated", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed.")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 9 Setup a linked internet gateway to prepare for next step
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
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.test", "id", func(v string) error {
						resourceId = v
						return nil
					}),
				),
			},
			// 10 - Unlink and link internet gateway to a new vpc
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
							return fmt.Errorf("Id should have changed.")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 11 - Delete vpc and link internet gateway to a new vpc
			{
				Config: `
resource "numspot_vpc" "test_even_more_new" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-internetgateway-acctest"
    }
  ]
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test_even_more_new.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_internet_gateway.test", "vpc_id", "numspot_vpc.test_even_more_new", "id"),
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
							return fmt.Errorf("Id should have changed.")
						}
						resourceId = v
						return nil
					}),
				),
			},
		},
	})
}
