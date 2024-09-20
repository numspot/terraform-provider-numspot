package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVolumeResource(t *testing.T) {
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
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
  link_vm = {
    vm_id       = numspot_vm.test.id
    device_name = "/dev/sdb"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.test", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.test", "size", "11"),
					resource.TestCheckResourceAttr("numspot_volume.test", "link_vm.device_name", "/dev/sdb"),
					resource.TestCheckResourceAttr("numspot_volume.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttrPair("numspot_volume.test", "link_vm.vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_volume.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
					}),
				),
				ExpectNonEmptyPlan: true,
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_volume.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_volume" "test" {
  type                   = "gp2"
  size                   = 22
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
  link_vm = {
    vm_id       = numspot_vm.test.id
    device_name = "/dev/sdc"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.test", "type", "gp2"),
					resource.TestCheckResourceAttr("numspot_volume.test", "size", "22"),
					resource.TestCheckResourceAttr("numspot_volume.test", "link_vm.device_name", "/dev/sdc"),
					resource.TestCheckResourceAttr("numspot_volume.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_volume.test", "link_vm.vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_volume.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
				ExpectNonEmptyPlan: true,
			},
			// 4 - Update testing With Replace of dependency resource and without Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test_new" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_volume" "test" {
  type                   = "gp2"
  size                   = 22
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
  link_vm = {
    vm_id       = numspot_vm.test_new.id
    device_name = "/dev/sdc"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.test", "type", "gp2"),
					resource.TestCheckResourceAttr("numspot_volume.test", "size", "22"),
					resource.TestCheckResourceAttr("numspot_volume.test", "link_vm.device_name", "/dev/sdc"),
					resource.TestCheckResourceAttr("numspot_volume.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_volume.test", "link_vm.vm_id", "numspot_vm.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_volume.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
				ExpectNonEmptyPlan: true,
			},
			// 5 - Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
			{
				Config: `
resource "numspot_volume" "test" {
  type                   = "gp2"
  size                   = 22
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.test", "type", "gp2"),
					resource.TestCheckResourceAttr("numspot_volume.test", "size", "22"),
					resource.TestCheckResourceAttr("numspot_volume.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_volume.test", "id", func(v string) error {
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
		},
	})
}
