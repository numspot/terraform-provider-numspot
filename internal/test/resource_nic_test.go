package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// 1 - Create unlinked Nic
// 2 - Import
// 3 - Update unlinked Nic
// 4 - Replace unlinked Nic
// 5 - Recreate unlinked Nic
// 6 - Link unlinked Nic (replace) (link to a VM and associate with security group)
//
// 7 - Update linked Nic
// 8 - Replace linked Nic
// 9 - Recreate linked Nic
// 10 - Unlink linked Nic
//
// 11 - //
// 12 - Unlink and link Nic to a new VM with deletion of old VM & security account

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
			{ // 1 - Create unlinked Nic
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "A beautiful Nic"
  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.50"
    },
    {
      is_primary = false
      private_ip = "10.101.1.100"
    }
  ]

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", "A beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test", "private_ips.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.50",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "false",
						"private_ip": "10.101.1.100",
					}),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			{ // 2 - ImportState testing
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"link_nic.state"},
			},
			{ // 3 - Update unlinked Nic
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "An even more beautiful Nic"
  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.50"
    },
    {
      is_primary = false
      private_ip = "10.101.1.100"
    }
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", "An even more beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test", "private_ips.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.50",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "false",
						"private_ip": "10.101.1.100",
					}),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						// TODO : For some reason, a replace is done here => to investigate
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 4 - Replace unlinked Nic
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "An even more beautiful Nic"
  private_ips = [
    {
      is_primary = false
      private_ip = "10.101.1.60"
    },
    {
      is_primary = true
      private_ip = "10.101.1.70"
    },
    {
      is_primary = false
      private_ip = "10.101.1.20"
    }
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", "An even more beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test", "private_ips.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "false",
						"private_ip": "10.101.1.60",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.70",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "false",
						"private_ip": "10.101.1.20",
					}),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 5 - Recreate unlinked Nic
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test_recreated" {
  subnet_id   = numspot_subnet.subnet.id
  description = "An even more beautiful Nic"
  private_ips = [
    {
      is_primary = false
      private_ip = "10.101.1.60"
    },
    {
      is_primary = true
      private_ip = "10.101.1.70"
    }
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test_recreated", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test_recreated", "description", "An even more beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test_recreated", "private_ips.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test_recreated", "private_ips.*", map[string]string{
						"is_primary": "false",
						"private_ip": "10.101.1.60",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test_recreated", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.70",
					}),
					resource.TestCheckResourceAttr("numspot_nic.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test_recreated", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 6 - Link unlinked Nic (replace) (link to a VM and associate with security group)
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
}

resource "numspot_vm" "vm" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "A beautiful Nic"

  security_group_ids = [numspot_security_group.test.id]

  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.60"
    },
    {
      is_primary = false
      private_ip = "10.101.1.70"
    }
  ]
  link_nic = {
    device_number = 1
    vm_id         = numspot_vm.vm.id
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", "A beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test", "private_ips.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.60",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "false",
						"private_ip": "10.101.1.70",
					}),
					resource.TestCheckResourceAttrPair("numspot_nic.test", "link_nic.vm_id", "numspot_vm.vm", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "link_nic.device_number", "1"),
					resource.TestCheckResourceAttr("numspot_nic.test", "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("numspot_nic.test", "security_group_ids.*", "numspot_security_group.test", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			{ // 7 - Update linked Nic
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
}

resource "numspot_vm" "vm" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "A more beautiful Nic"

  security_group_ids = [numspot_security_group.test.id]

  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.60"
    },
    {
      is_primary = false
      private_ip = "10.101.1.70"
    }
  ]
  link_nic = {
    device_number = 1
    vm_id         = numspot_vm.vm.id
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", "A more beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test", "private_ips.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.60",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "false",
						"private_ip": "10.101.1.70",
					}),
					resource.TestCheckResourceAttrPair("numspot_nic.test", "link_nic.vm_id", "numspot_vm.vm", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "link_nic.device_number", "1"),
					resource.TestCheckResourceAttr("numspot_nic.test", "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("numspot_nic.test", "security_group_ids.*", "numspot_security_group.test", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 8 - Replace linked Nic
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
}

resource "numspot_vm" "vm" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "A more beautiful Nic"

  security_group_ids = [numspot_security_group.test.id]

  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.70"
    }
  ]
  link_nic = {
    device_number = 1
    vm_id         = numspot_vm.vm.id
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", "A more beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test", "private_ips.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.70",
					}),
					resource.TestCheckResourceAttrPair("numspot_nic.test", "link_nic.vm_id", "numspot_vm.vm", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "link_nic.device_number", "1"),
					resource.TestCheckResourceAttr("numspot_nic.test", "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("numspot_nic.test", "security_group_ids.*", "numspot_security_group.test", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 9 - Recreate linked Nic
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
}

resource "numspot_vm" "vm" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_nic" "test_recreated" {
  subnet_id   = numspot_subnet.subnet.id
  description = "A more beautiful Nic"

  security_group_ids = [numspot_security_group.test.id]

  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.70"
    }
  ]
  link_nic = {
    device_number = 1
    vm_id         = numspot_vm.vm.id
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test_recreated", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test_recreated", "description", "A more beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test_recreated", "private_ips.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test_recreated", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.70",
					}),
					resource.TestCheckResourceAttrPair("numspot_nic.test_recreated", "link_nic.vm_id", "numspot_vm.vm", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test_recreated", "link_nic.device_number", "1"),
					resource.TestCheckResourceAttr("numspot_nic.test_recreated", "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("numspot_nic.test_recreated", "security_group_ids.*", "numspot_security_group.test", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test_recreated", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 10 - Unlink linked Nic
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "A more beautiful Nic"

  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.70"
    }
  ]

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", "A more beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test", "private_ips.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.70",
					}),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC-Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 11 - Setup for next step
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
}

resource "numspot_vm" "vm" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "A beautiful Nic"

  security_group_ids = [numspot_security_group.test.id]

  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.70"
    }
  ]
  link_nic = {
    device_number = 1
    vm_id         = numspot_vm.vm.id
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC"
    }
  ]
}`,
			},
			{ // 12 - Unlink and link Nic to a new VM with deletion of old VM & security account
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test_new" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
}

resource "numspot_vm" "vm_new" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.subnet.id
  description = "A beautiful Nic"

  security_group_ids = [numspot_security_group.test_new.id]

  private_ips = [
    {
      is_primary = true
      private_ip = "10.101.1.70"
    }
  ]
  link_nic = {
    device_number = 1
    vm_id         = numspot_vm.vm_new.id
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-NIC"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.subnet", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", "A beautiful Nic"),
					resource.TestCheckResourceAttr("numspot_nic.test", "private_ips.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "private_ips.*", map[string]string{
						"is_primary": "true",
						"private_ip": "10.101.1.70",
					}),
					resource.TestCheckResourceAttrPair("numspot_nic.test", "link_nic.vm_id", "numspot_vm.vm_new", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "link_nic.device_number", "1"),
					resource.TestCheckResourceAttr("numspot_nic.test", "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("numspot_nic.test", "security_group_ids.*", "numspot_security_group.test_new", "id"),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-NIC",
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
