//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	vmType        = "tinav6.c1r1p3"
	sourceImageId = "ami-026ce760"
)

func TestAccVmResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_Create(sourceImageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
				ExpectNonEmptyPlan: true,
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
				Config:             testVmConfig_Create(sourceImageId, vmType),
				Check:              resource.ComposeAggregateTestCheckFunc(),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testVmConfig_Create(sourceImageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_image" "test" {
	name               = "terraform-generated-image-for-public-ip-test"
	source_image_id    = %[1]q
	source_region_name = "cloudgouv-eu-west-1"
}
resource "numspot_vm" "test" {
  image_id = numspot_image.test.id
  vm_type  = %[2]q
}
`, sourceImageId, vmType)
}

func TestAccVmResource_NetSubnet(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_NetSubnet(sourceImageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "net_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
				),
				ExpectNonEmptyPlan: true,
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
				Config:             testVmConfig_NetSubnet(sourceImageId, vmType),
				Check:              resource.ComposeAggregateTestCheckFunc(),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testVmConfig_NetSubnet(sourceImageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_image" "test" {
	name               = "terraform-generated-image-for-public-ip-test"
	source_image_id    = %[1]q
	source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_vm" "test" {
  image_id  = numspot_image.test.id
  vm_type   = %[2]q
  subnet_id = numspot_subnet.subnet.id
}
`, sourceImageId, vmType)
}

func TestAccVmResource_NetSubnetSG(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_NetSubnetSG(sourceImageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "net_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
				),
			},
		},
	})
}

func testVmConfig_NetSubnetSG(sourceImageId, vmType string) string {
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

resource "numspot_image" "test" {
	name               = "terraform-generated-image-for-vm-test"
	source_image_id    = %[1]q
	source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_vm" "test" {
  image_id           = numspot_image.test.id
  vm_type            = %[2]q
  subnet_id          = numspot_subnet.subnet.id
  security_group_ids = [numspot_security_group.sg.id]
  depends_on         = [numspot_security_group.sg]
}
`, sourceImageId, vmType)
}

func TestAccVmResource_NetSubnetSGRouteTable(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_NetSubnetSGRouteTable(sourceImageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "net_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testVmConfig_NetSubnetSGRouteTable(sourceImageId, vmType string) string {
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

resource "numspot_image" "test" {
	name               = "terraform-generated-image-for-vm-test"
	source_image_id    = %[1]q
	source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_vm" "test" {
  image_id           = numspot_image.test.id
  vm_type            = %[2]q
  subnet_id          = numspot_subnet.subnet.id
  security_group_ids = [numspot_security_group.sg.id]
  depends_on         = [numspot_security_group.sg]
}
`, sourceImageId, vmType)
}

func TestAccVmResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	tagKey := "Name"
	tagValue := "terraform-vm"
	tagValueUpdated := tagValue + "-Updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_Tags(sourceImageId, vmType, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1")),
				ExpectNonEmptyPlan: true,
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
				Config: testVmConfig_Tags(sourceImageId, vmType, tagKey, tagValueUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.0.value", tagValueUpdated),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testVmConfig_Tags(imageId, vmType, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_vm" "test" {
  image_id = %[1]q
  vm_type  = %[2]q

  tags = [
	{
	  key 		= %[3]q
	  value	 	= %[4]q
	}
  ]
}
`, imageId, vmType, tagKey, tagValue)
}
