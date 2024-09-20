package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccNicResource(t *testing.T) {
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
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
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
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
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
			// 4 - Update testing With Replace of VPC and without Replacing Internet Gateway (if needed)
			// This test is useful to check wether or not the deletion of the VPC and then the update of the Internet Gateway works properly
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet_new" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet_new.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet_new", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
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
