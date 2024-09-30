package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVmResource(t *testing.T) {
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
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.net.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = "terraform-vm-tests-sg-name"
  description = "terraform-vm-tests-sg-description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_vm" "test" {
  image_id                       = "ami-0987a84b"
  type                           = "ns-eco6-2c8r"
  vm_initiated_shutdown_behavior = "stop"


  tags = [
    {
      key   = "name"
      value = "Terraform-Test-VM"
    }
  ]

  placement = {
    tenancy                = "default"
    availability_zone_name = "cloudgouv-eu-west-1a"
  }

  subnet_id          = numspot_subnet.test.id
  security_group_ids = [numspot_security_group.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "image_id", "ami-0987a84b"),
					resource.TestCheckResourceAttr("numspot_vm.test", "type", "ns-eco6-2c8r"),
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_initiated_shutdown_behavior", "stop"),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-VM",
					}),
					resource.TestCheckResourceAttr("numspot_vm.test", "placement.availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttrPair("numspot_vm.test", "subnet_id", "numspot_subnet.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_vm.test", "security_group_ids.*", "numspot_security_group.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
					}),
				),
			},
			{ // 2 - ImportState testing
				ResourceName:            "numspot_vm.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace
			{
				Config: `
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.net.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = "terraform-vm-tests-sg-name"
  description = "terraform-vm-tests-sg-description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_vm" "test" {
  image_id                       = "ami-0987a84b"
  type                           = "ns-cus6-4c8r"
  vm_initiated_shutdown_behavior = "stop"


  tags = [
    {
      key   = "name"
      value = "Terraform-Test-VM-Updated"
    }
  ]

  placement = {
    tenancy                = "default"
    availability_zone_name = "cloudgouv-eu-west-1a"
  }

  subnet_id          = numspot_subnet.test.id
  security_group_ids = [numspot_security_group.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "image_id", "ami-0987a84b"),
					resource.TestCheckResourceAttr("numspot_vm.test", "type", "ns-cus6-4c8r"),
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_initiated_shutdown_behavior", "stop"),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-VM-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_vm.test", "placement.availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttrPair("numspot_vm.test", "subnet_id", "numspot_subnet.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_vm.test", "security_group_ids.*", "numspot_security_group.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
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
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test_new" {
  vpc_id                 = numspot_vpc.net.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_security_group" "test_new" {
  vpc_id      = numspot_vpc.net.id
  name        = "terraform-vm-tests-sg-name"
  description = "terraform-vm-tests-sg-description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_vm" "test" {
  image_id                       = "ami-0987a84b"
  type                           = "ns-cus6-4c8r"
  vm_initiated_shutdown_behavior = "stop"


  tags = [
    {
      key   = "name"
      value = "Terraform-Test-VM-Updated"
    }
  ]

  placement = {
    tenancy                = "default"
    availability_zone_name = "cloudgouv-eu-west-1a"
  }

  subnet_id          = numspot_subnet.test_new.id
  security_group_ids = [numspot_security_group.test_new.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "image_id", "ami-0987a84b"),
					resource.TestCheckResourceAttr("numspot_vm.test", "type", "ns-cus6-4c8r"),
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_initiated_shutdown_behavior", "stop"),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-VM-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_vm.test", "placement.availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttrPair("numspot_vm.test", "subnet_id", "numspot_subnet.test_new", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_vm.test", "security_group_ids.*", "numspot_security_group.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
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

func TestAccVmResource_NetSubnetSGRouteTable(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: `
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.net.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = "terraform-vm-tests-sg-name"
  description = "terraform-vm-tests-sg-description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_internet_gateway" "igw" {
  vpc_id = numspot_vpc.net.id
}

resource "numspot_route_table" "rt" {
  vpc_id    = numspot_vpc.net.id
  subnet_id = numspot_subnet.test.id

  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.igw.id
    }
  ]
}

resource "numspot_public_ip" "public_ip" {
  vm_id      = numspot_vm.test.id
  depends_on = [numspot_route_table.rt]
}

resource "numspot_vm" "test" {
  image_id           = "ami-0987a84b"
  type               = "ns-eco6-2c8r"
  subnet_id          = numspot_subnet.test.id
  security_group_ids = [numspot_security_group.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "type", "ns-eco6-2c8r"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "vpc_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_vm.test", "security_group_ids.*", "numspot_security_group.test", "id"),
					resource.TestCheckResourceAttrPair("numspot_vm.test", "subnet_id", "numspot_subnet.test", "id"),
				),
			},
		},
	})
}
