//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// 1 - Create Image from Image (public)
// 2 - Import
// 3 - Replace Image from Image (public)
// 4 - Recreate Image form Image (public)
// 5 - Update Image from Image (private)
//
// 6 - Replace Image from VM (public)
// 7 - Recreate Image form VM (public)
// 8 - Update Image from VM (private)
//
// 9 - Replace Image from Snapshot (public)
// 10 - Recreate Image from Snapshot (public)
// 11 - Update Image from Snapshot (private)
//
// 12 - Reset state to Image from VM - prepare next step (public)
// 13 - Replace Image with a new VM (public)
//
// 14 - Reset state to Image from Snapshot - prepare next step (public)
// 15 - Replace Image with a new Snapshot (private)

func TestAccImageResource(t *testing.T) {
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
			{ // 1 - Create Image from Image
				Config: `
resource "numspot_image" "test" {
  name               = "terraform image test v2"
  source_image_id    = "ami-88dbc914"
  source_region_name = "eu-west-2"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform image test v2"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", "ami-88dbc914"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", "eu-west-2"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			{ // 2 - ImportState testing
				ResourceName:            "numspot_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_image_id", "source_region_name"},
			},
			{ // 3 - Replace Image from Image
				Config: `
resource "numspot_image" "test" {
  name               = "terraform image test v2 updated"
  source_image_id    = "ami-88dbc914"
  source_region_name = "eu-west-2"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", "ami-88dbc914"),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", "eu-west-2"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 4 - Recreate Image from Image
				Config: `
resource "numspot_image" "test_recreate" {
  name               = "terraform image test v2 updated"
  source_image_id    = "ami-88dbc914"
  source_region_name = "eu-west-2"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "source_image_id", "ami-88dbc914"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "source_region_name", "eu-west-2"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 5 - Update Image from Image
				Config: `
resource "numspot_image" "test_recreate" {
  name               = "terraform image test v2 updated"
  source_image_id    = "ami-88dbc914"
  source_region_name = "eu-west-2"
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "source_image_id", "ami-88dbc914"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "source_region_name", "eu-west-2"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 6 - Replace Image from VM
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
}
resource "numspot_vm" "test" {
  image_id  = "ami-88dbc914"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = "terraform image test v2 updated"
  vm_id = numspot_vm.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_image.test", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 7 - Recreate Image from Vm
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
}
resource "numspot_vm" "test" {
  image_id  = "ami-88dbc914"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test_recreate" {
  name  = "terraform image test v2 updated"
  vm_id = numspot_vm.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_image.test_recreate", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 8 - Update Image from VM
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
}
resource "numspot_vm" "test" {
  image_id  = "ami-88dbc914"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test_recreate" {
  name  = "terraform image test v2 updated"
  vm_id = numspot_vm.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_image.test_recreate", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 9 - Replace Image from Snapshot
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2b"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "a numspot snapshot description"
}

resource "numspot_image" "test" {
  name             = "terraform image test v2 updated"
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
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test", "block_device_mappings.*", map[string]string{
						"bsu.snapshot_id": acctest.PairPrefix + "numspot_snapshot.test.id",
					}),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 10 - Recreate Image from Snapshot
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2b"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "a numspot snapshot description"
}

resource "numspot_image" "test_recreate" {
  name             = "terraform image test v2 updated"
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
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test_recreate", "block_device_mappings.*", map[string]string{
						"bsu.snapshot_id": acctest.PairPrefix + "numspot_snapshot.test.id",
					}),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 11 - Update Image from Snapshot
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2b"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "a numspot snapshot description"
}

resource "numspot_image" "test_recreate" {
  name             = "terraform image test v2 updated"
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
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test_recreate", "block_device_mappings.*", map[string]string{
						"bsu.snapshot_id": acctest.PairPrefix + "numspot_snapshot.test.id",
					}),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test_recreate", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 12 - Reset state to Image from VM to prepare next step
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
}
resource "numspot_vm" "test" {
  image_id  = "ami-88dbc914"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = "terraform image test v2"
  vm_id = numspot_vm.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
			},
			{ // 13 - Replace Image with a new VM -  This test is useful to check wether or not the deletion of VM and then the deletion of the Image works properly
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
}
resource "numspot_vm" "test_new" {
  image_id  = "ami-88dbc914"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = "terraform image test v2 updated"
  vm_id = numspot_vm.test_new.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_image.test", "vm_id", "numspot_vm.test_new", "id"),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 14 - Reset state to Image from Snapshot - prepare next step
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
}
resource "numspot_vm" "test" {
  image_id  = "ami-88dbc914"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = "terraform image test v2"
  vm_id = numspot_vm.test.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
			},
			{ // 15 - Update testing With Replace of Snapshot and with Replace of the Image - This test is useful to check wether or not the deletion of the Snapshot and then the deletion of the Image works properly
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2b"
}

resource "numspot_snapshot" "test_new" {
  volume_id   = numspot_volume.test.id
  description = "a numspot snapshot description"
}

resource "numspot_image" "test" {
  name             = "terraform image test v2 updated"
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
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Image"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test", "block_device_mappings.*", map[string]string{
						"bsu.snapshot_id": acctest.PairPrefix + "numspot_snapshot.test_new.id",
					}),
					resource.TestCheckResourceAttr("numspot_image.test", "name", "terraform image test v2 updated"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Image",
					}),
					resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
