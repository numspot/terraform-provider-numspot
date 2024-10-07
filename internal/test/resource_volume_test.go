package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// Create unlinked volume
// Update attributes from unlinked volume
// Replace attributes // no replace
// Recreate unlinked volume
// Link unlinked volume
//
// Create linked volume
// Update attributes from linked volume
// Replace attributes // no replace
// Unlink linked volume
// Recreate linked volume
//
// Unlink and link volume to a new VM
// Delete VM and link volume to a new VM

func TestAccVolumeResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	volumeDependencies := `
resource "numspot_vpc" "terraform-dep-vpc-volume" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-dep-vpc-volume"
    }
  ]
}

resource "numspot_subnet" "terraform-dep-subnet-volume" {
  vpc_id                 = numspot_vpc.terraform-dep-vpc-volume.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "terraform-dep-subnet-volume"
    }
  ]
}

resource "numspot_vm" "terraform-dep-vm-volume" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.terraform-dep-subnet-volume.id
  tags = [
    {
      key   = "name"
      value = "terraform-dep-vm-volume"
    }
  ]
}
		`

	volumeUpdateLinkDependencies := `
resource "numspot_vpc" "terraform-dep-vpc-volume" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-dep-vpc-volume"
    }
  ]
}

resource "numspot_subnet" "terraform-dep-subnet-volume" {
  vpc_id                 = numspot_vpc.terraform-dep-vpc-volume.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "terraform-dep-subnet-volume"
    }
  ]
}

resource "numspot_vm" "terraform-dep-vm-volume" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.terraform-dep-subnet-volume.id
  tags = [
    {
      key   = "name"
      value = "terraform-dep-vm-volume"
    }
  ]
}

resource "numspot_vm" "terraform-dep-vm-volume-dest-link" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.terraform-dep-subnet-volume.id
  tags = [
    {
      key   = "name"
      value = "terraform-dep-vm-volume-dest-link"
    }
  ]
}
			`
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Step 1 - Create unlinked volume
			{
				Config: `
resource "numspot_volume" "terraform-volume-acctest" {
  type                   = "standard"
  size                   = 10
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}
										`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "size", "10"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest",
					}),
				),
			},
			// Step 2 - Import
			{
				ResourceName:            "numspot_volume.terraform-volume-acctest",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Step 3 - Update attributes from unlinked volume
			{
				Config: `
resource "numspot_volume" "terraform-volume-acctest" {
  type                   = "gp2"
  size                   = 15
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest-update"
    }
  ]
}
							`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "type", "gp2"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "size", "15"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest-update",
					}),
				),
			},
			// Step 4 - Recreate unlinked volume
			{
				Config: `
resource "numspot_volume" "terraform-volume-acctest-recreate" {
  type                   = "standard"
  size                   = 10
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest-recreate"
    }
  ]
}
							`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "size", "10"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest-recreate",
					}),
				),
			},
			// Step 5 - Link unlinked volume
			{
				Config: volumeDependencies + `
resource "numspot_volume" "terraform-volume-acctest-recreate" {
  type                   = "standard"
  size                   = 10
  availability_zone_name = "cloudgouv-eu-west-1a"
  link_vm = {
    vm_id       = numspot_vm.terraform-dep-vm-volume.id
    device_name = "/dev/sdb"
  }
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest-recreate"
    }
  ]
}
										`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "size", "10"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "tags.#", "1"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "link_vm.device_name", "/dev/sdb"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest-recreate",
					}),
				),
			},
			// Step 6 - Unlink linked volume
			{
				Config: volumeDependencies + `
resource "numspot_volume" "terraform-volume-acctest-recreate" {
  type                   = "standard"
  size                   = 10
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest-recreate"
    }
  ]
}
										`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "size", "10"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest-recreate",
					}),
				),
			},
			// Step 7 - Delete unlinked volume
			{
				Config: volumeDependencies,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Step 8 - Create linked volume
			{
				Config: volumeDependencies + `
resource "numspot_volume" "terraform-volume-acctest" {
  type                   = "standard"
  size                   = 10
  availability_zone_name = "cloudgouv-eu-west-1a"
  link_vm = {
    vm_id       = numspot_vm.terraform-dep-vm-volume.id
    device_name = "/dev/sdb"
  }
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}
										`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "size", "10"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "tags.#", "1"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "link_vm.device_name", "/dev/sdb"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest",
					}),
				),
			},
			// Step 9 - Update attributes from linked volume
			{
				Config: volumeDependencies + `
resource "numspot_volume" "terraform-volume-acctest" {
  type                   = "gp2"
  size                   = 15
  availability_zone_name = "cloudgouv-eu-west-1a"
  link_vm = {
    vm_id       = numspot_vm.terraform-dep-vm-volume.id
    device_name = "/dev/sdc"
  }
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest-update"
    }
  ]
}
										`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "type", "gp2"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "size", "15"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "tags.#", "1"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest", "link_vm.device_name", "/dev/sdc"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest-update",
					}),
				),
			},
			// Edge case on linked volume recreation, since Terraform triggers concurrently Create and Delete when recreating a resource,  i.e. when a resource name changes
			// Create can be called before Delete, and we will try to link a different volume to the same VM and device (which returns a conflict and won't work without a retry link)
			// Step 10 - Recreate linked volume
			{
				Config: volumeDependencies + `
resource "numspot_volume" "terraform-volume-acctest-recreate" {
  type                   = "standard"
  size                   = 20
  availability_zone_name = "cloudgouv-eu-west-1a"
  link_vm = {
    vm_id       = numspot_vm.terraform-dep-vm-volume.id
    device_name = "/dev/sdc"
  }
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest-recreate"
    }
  ]
}
										`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "size", "20"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "tags.#", "1"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "link_vm.device_name", "/dev/sdc"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest-recreate",
					}),
				),
			},
			// Step 11 - Unlink volume and link to a new VM
			{
				Config: volumeUpdateLinkDependencies + `
resource "numspot_volume" "terraform-volume-acctest-recreate" {
  type                   = "standard"
  size                   = "20"
  availability_zone_name = "cloudgouv-eu-west-1a"
  link_vm = {
    vm_id       = numspot_vm.terraform-dep-vm-volume-dest-link.id
    device_name = "/dev/sdc"
  }
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest-recreate"
    }
  ]
}
							`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "size", "20"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "tags.#", "1"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "link_vm.device_name", "/dev/sdc"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest-recreate",
					}),
				),
			},
			// Edge case on unlinking a volume after VM removal
			// Default unlinking behavior includes stopping the currently linked VM (which should not exist anymore in this case)
			// Step 12 - Unlink by removing VM
			{
				Config: `
resource "numspot_volume" "terraform-volume-acctest-recreate" {
  type                   = "standard"
  size                   = 25
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest-recreate"
    }
  ]
}
							`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "type", "standard"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "size", "25"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttr("numspot_volume.terraform-volume-acctest-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.terraform-volume-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-volume-acctest-recreate",
					}),
				),
			},
		},
	})
}
