//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	imageId = "ami-00b0c39a"
	vmType  = "t2.small"
)

func TestAccVmResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_Create(imageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "image_id", imageId),
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vm.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testVmConfig_Create(imageId, vmType),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testVmConfig_Create(imageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_vm" "test" {
  image_id = %[1]q
  vm_type  = %[2]q
}
`, imageId, vmType)
}

func TestAccVmResource_NetSubnet(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_NetSubnet(imageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "image_id", imageId),
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "net_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vm.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testVmConfig_NetSubnet(imageId, vmType),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testVmConfig_NetSubnet(imageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_vm" "test" {
  image_id  = %[1]q
  vm_type   = %[2]q
  subnet_id = numspot_subnet.subnet.id
}
`, imageId, vmType)
}

func TestAccVmResource_NetSubnetSG(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_NetSubnetSG(imageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "image_id", imageId),
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "net_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
				),
			},
		},
	})
}

func testVmConfig_NetSubnetSG(imageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "sg" {
  net_id      = numspot_vpc.net.id
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
  image_id           = %[1]q
  vm_type            = %[2]q
  subnet_id          = numspot_subnet.subnet.id
  security_group_ids = [numspot_security_group.sg.id]
  depends_on         = [numspot_security_group.sg]
}
`, imageId, vmType)
}

func TestAccVmResource_NetSubnetSGRouteTable(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_NetSubnetSGRouteTable(imageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "image_id", imageId),
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "net_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
				),
			},
		},
	})
}

func testVmConfig_NetSubnetSGRouteTable(imageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "sg" {
  net_id      = numspot_vpc.net.id
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
  net_id = numspot_vpc.net.id
}

resource "numspot_route_table" "rt" {
  net_id    = numspot_vpc.net.id
  subnet_id = numspot_subnet.subnet.id

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
  image_id           = %[1]q
  vm_type            = %[2]q
  subnet_id          = numspot_subnet.subnet.id
  security_group_ids = [numspot_security_group.sg.id]
  depends_on         = [numspot_security_group.sg]
}
`, imageId, vmType)
}
