package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// 1 - Create unlinked PublicIP
// 2 - Import
// 3 - Update attributes from unlinked PublicIP
// 4 - Recreate unlinked PublicIP
//
// 5 - Link unlinked PublicIP to VM
// 6 - Update attributes from linked PublicIP (to VM)
// 7 - Recreate linked PublicIP (to VM)
// 8 - Unlink linked PublicIP (to VM)
//
// 9 - Link unlinked PublicIP to a NIC
// 10 - Update attributes from linked PublicIP (to NIC)
// 11 - Recreate linked PublicIP (to NIC)
// 12 - Unlink linked PublicIP (to NIC)
//
// 13 - setup for next step
// 14 - Unlink and link PublicIP from a NIC to a new NIC
// 15 - Delete NIC and link PublicIP to a new NIC
// 16 - setup for next step
// 17 - Unlink and link PublicIP from a VM to a new VM
// 18 - Delete VM and link PublicIP to a new VM

func TestAccPublicIpResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create unlinked PublicIP
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
				),
			},
			{ // 2 - ImportState testing
				ResourceName:            "numspot_public_ip.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{ // 3 - Update attributes from unlinked PublicIP
				Config: `
resource "numspot_public_ip" "test" {
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp-Updated",
					}),
				),
			},
			{ // 4 - Recreate unlinked PublicIP
				Config: `
resource "numspot_public_ip" "test_recreated" {
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_public_ip.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
				),
			},
			{ // 5 - Link unlinked PublicIP to VM
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
				),
			},
			{ // 6 - Update attributes from linked PublicIP (to VM)
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
				),
			},
			{ // 7 - Recreate linked PublicIP (to VM)
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

resource "numspot_public_ip" "test_recreated" {
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
					resource.TestCheckResourceAttrPair("numspot_public_ip.test_recreated", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
				),
			},
			{ // 8 - Unlink linked PublicIP (to VM)
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
				),
			},
			{ // 9 - Link unlinked PublicIP to a NIC
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

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  nic_id     = numspot_nic.test.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "nic_id", "numspot_nic.test", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
				),
			},
			{ // 10 - Update attributes from linked PublicIP (to NIC)
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

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
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
				),
			},
			{ // 11 - Recreate linked PublicIP (to NIC)
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

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test_recreated" {
  nic_id     = numspot_nic.test.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test_recreated", "nic_id", "numspot_nic.test", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
				),
			},
			{ // 12 - Unlink linked PublicIP (to NIC)
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
				),
			},
			{ // 13 - setup for next step
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

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  nic_id     = numspot_nic.test.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
			},
			{ // 14 - Unlink and link PublicIP from a NIC to a new NIC
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

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.test.id
}

resource "numspot_nic" "test_new" {
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  nic_id     = numspot_nic.test_new.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "nic_id", "numspot_nic.test_new", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
				),
			},
			{ // 15 - Delete NIC and link PublicIP to a new NIC
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

resource "numspot_nic" "test_newest" {
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  nic_id     = numspot_nic.test_newest.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "nic_id", "numspot_nic.test_newest", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
				),
			},
			{ // 16 - setup for next step
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
			},
			{ // 17 - Unlink and link PublicIP from a VM to a new VM
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
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "vm_id", "numspot_vm.test_new", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
				),
			},
			{ // 18 - Delete VM and link PublicIP to a new VM
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

resource "numspot_vm" "test_newest" {
  image_id  = numspot_image.test.id
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  vm_id      = numspot_vm.test_newest.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "vm_id", "numspot_vm.test_newest", "id"),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-PublicIp",
					}),
				),
			},
		},
	})
}
