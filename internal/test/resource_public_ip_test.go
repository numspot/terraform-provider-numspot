package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccPublicIpResource(t *testing.T) {
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
			{ // 1 - Create testing with Vm
				Config: `
resource "numspot_image" "test" {
  name               = "terraform-generated-image-for-public-ip-acctest"
  source_image_id    = "ami-0b7df82c"
  source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test" {
  image_id  = numspot_image.test.id
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  vm_id      = numspot_vm.test.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "id", func(v string) error {
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
				ResourceName:            "numspot_public_ip.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace with Vm
			{
				Config: `
resource "numspot_image" "test" {
  name               = "terraform-generated-image-for-public-ip-acctest"
  source_image_id    = "ami-0b7df82c"
  source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test" {
  image_id  = numspot_image.test.id
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  vm_id      = numspot_vm.test.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "id", func(v string) error {
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
			// 4 - Update testing With Replace of dependency resource and with Replace of the resource
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
			{
				Config: `
resource "numspot_image" "test" {
  name               = "terraform-generated-image-for-public-ip-acctest"
  source_image_id    = "ami-0b7df82c"
  source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test_new" {
  image_id  = numspot_image.test.id
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  vm_id      = numspot_vm.test_new.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "vm_id", "numspot_vm.test_new", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						resourceId = v
						return nil
					}),
				),
			},
			{ // 5 - Update testing With removal of dependency resource and with Replace of the resource

				Config: `
resource "numspot_public_ip" "test" {
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "id", func(v string) error {
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
			{ // 6 - Replace public ip and create from NIC
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
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_public_ip" "test" {
  nic_id     = numspot_nic.test.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "nic_id", "numspot_nic.test", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						resourceId = v
						return nil
					}),
				),
			},
		},
	})
}
