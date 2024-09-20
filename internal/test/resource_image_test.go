package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccImageResource(t *testing.T) {
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
resource "numspot_image" "test" {
  name               = "terraform-image-test"
  source_image_id    = "ami-0b7df82c"
  source_region_name = "cloudgouv-eu-west-1"
  access = {
    is_public = "true"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", "ami-0b7df82c"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", "cloudgouv-eu-west-1"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "true"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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
				ResourceName:            "numspot_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_image_id", "source_region_name"},
			},
			// 3 - Update testing With Replace (create image from Image)
			{
				Config: `
resource "numspot_image" "test" {
  name               = "terraform-image-test-updated"
  source_image_id    = "ami-0987a84b"
  source_region_name = "cloudgouv-eu-west-1"
  access = {
    is_public = "true"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test-updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", "ami-0987a84b"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", "cloudgouv-eu-west-1"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "true"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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
			// 4 - Update testing Without Replace (create image from Image)
			{
				Config: `
resource "numspot_image" "test" {
  name               = "terraform-image-test-updated"
  source_image_id    = "ami-0987a84b"
  source_region_name = "cloudgouv-eu-west-1"
  access = {
    is_public = "false"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test-updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", "ami-0987a84b"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", "cloudgouv-eu-west-1"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "false"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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

			// 5 - Update testing With Replace (create image from Vm)
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}
resource "numspot_vm" "test" {
  image_id  = "ami-0987a84b"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = "terraform-image-test-updated"
  vm_id = numspot_vm.test.id
  access = {
    is_public = "true"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_image.test", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test-updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "true"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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
			// 6 - Update testing Without Replace (create image from Vm)
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}
resource "numspot_vm" "test" {
  image_id  = "ami-0987a84b"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = "terraform-image-test-updated"
  vm_id = numspot_vm.test.id
  access = {
    is_public = "false"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_image.test", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test-updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "false"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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

			// 7 - Update testing With Replace (create image from Snapshot)
			{
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1b"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "a numspot snapshot description"
}

resource "numspot_image" "test" {
  name             = "terraform-image-test-updated"
  root_device_name = "/dev/sda1"
  block_device_mappings = [
    {
      device_name = "/dev/sda1"
      bsu = {
        snapshot_id           = numspot_snapshot.test.id
        volume_size           = 120
        volume_type           = "io1"
        iops                  = 150
        delete_on_vm_deletion = true
      }
    }
  ]
  access = {
    is_public = "true"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test", "block_device_mappings.*", map[string]string{
						"bsu.snapshot_id": acctest.PAIR_PREFIX + "numspot_snapshot.test.id",
					}),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test-updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "true"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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
			// 8 - Update testing Without Replace (create image from Snapshot)
			{
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1b"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "a numspot snapshot description"
}

resource "numspot_image" "test" {
  name             = "terraform-image-test-updated"
  root_device_name = "/dev/sda1"
  block_device_mappings = [
    {
      device_name = "/dev/sda1"
      bsu = {
        snapshot_id           = numspot_snapshot.test.id
        volume_size           = 120
        volume_type           = "io1"
        iops                  = 150
        delete_on_vm_deletion = true
      }
    }
  ]
  access = {
    is_public = "false"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test", "block_device_mappings.*", map[string]string{
						"bsu.snapshot_id": acctest.PAIR_PREFIX + "numspot_snapshot.test.id",
					}),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test-updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "false"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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
			{ // 9 - Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: `
resource "numspot_image" "test" {
  name               = "terraform-image-test"
  source_image_id    = "ami-0b7df82c"
  source_region_name = "cloudgouv-eu-west-1"
  access = {
    is_public = "true"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
			},
			// 10 - Update testing With Replace of VM and with Replace of Image
			// This test is useful to check wether or not the deletion of VM and then the deletion of the Image works properly
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}
resource "numspot_vm" "test_new" {
  image_id  = "ami-0987a84b"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = "terraform-image-test-updated"
  vm_id = numspot_vm.test_new.id
  access = {
    is_public = "true"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_image.test", "vm_id", "numspot_vm.test_new", "id"),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test-updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "true"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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
			{ // 11 - Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}
resource "numspot_vm" "test" {
  image_id  = "ami-0987a84b"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = "terraform-image-test"
  vm_id = numspot_vm.test.id
  access = {
    is_public = "true"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
			},
			// 12 - Update testing With Replace of Snapshot and with Replace of the Image
			// This test is useful to check wether or not the deletion of the Snapshot and then the deletion of the Image works properly
			{
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1b"
}

resource "numspot_snapshot" "test_new" {
  volume_id   = numspot_volume.test.id
  description = "a numspot snapshot description"
}

resource "numspot_image" "test" {
  name             = "terraform-image-test-updated"
  root_device_name = "/dev/sda1"
  block_device_mappings = [
    {
      device_name = "/dev/sda1"
      bsu = {
        snapshot_id           = numspot_snapshot.test_new.id
        volume_size           = 120
        volume_type           = "io1"
        iops                  = 150
        delete_on_vm_deletion = true
      }
    }
  ]
  access = {
    is_public = "false"
  }
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test", "block_device_mappings.*", map[string]string{
						"bsu.snapshot_id": acctest.PAIR_PREFIX + "numspot_snapshot.test_new.id",
					}),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform-image-test-updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", "false"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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
