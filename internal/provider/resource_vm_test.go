//go:build acc

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	vmType        = "ns-cus6-2c4r"
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
resource "numspot_vm" "test" {
  image_id = %[1]q
  type     = %[2]q
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
					resource.TestCheckResourceAttr("numspot_vm.test", "type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "vpc_id"),
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

resource "numspot_vm" "test" {
  image_id  = %[1]q
  type      = %[2]q
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
					resource.TestCheckResourceAttr("numspot_vm.test", "type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "vpc_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
				),
				ExpectNonEmptyPlan: true,
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

resource "numspot_vm" "test" {
  image_id           = %[1]q
  type               = %[2]q
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
					resource.TestCheckResourceAttr("numspot_vm.test", "type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "vpc_id"),
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
  vpc_id = numspot_vpc.net.id
}

resource "numspot_route_table" "rt" {
  vpc_id    = numspot_vpc.net.id
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
  type               = %[2]q
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
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1"),
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
  type     = %[2]q

  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}
`, imageId, vmType, tagKey, tagValue)
}

func TestAccVmResource_Update_WithoutReplace(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	vmTypeUpdated := "ns-cus6-4c8r"
	vmInitiatedShutdownBehavior := "stop"
	vmInitiatedShutdownBehaviorUpdated := "terminate"

	tagKey := "Name"
	tagValue := "terraform-vm"
	tagValueUpdated := tagValue + "-Updated"

	performance := "medium"
	performanceUpdated := "high"
	var vm_id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_UpdateNoReplace(sourceImageId, vmType, vmInitiatedShutdownBehavior, tagKey, tagValue, performance),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						vm_id = v
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vm.test", "type", vmType),
					resource.TestCheckResourceAttr("numspot_vm.test", "performance", performance),
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_initiated_shutdown_behavior", vmInitiatedShutdownBehavior),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.0.value", tagValue),
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
				Config: testVmConfig_UpdateNoReplace(sourceImageId, vmTypeUpdated, vmInitiatedShutdownBehaviorUpdated, tagKey, tagValueUpdated, performanceUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						if vm_id != v {
							return errors.New("Id should be the same after Update without replace")
						}
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vm.test", "type", vmTypeUpdated),
					resource.TestCheckResourceAttr("numspot_vm.test", "vm_initiated_shutdown_behavior", vmInitiatedShutdownBehaviorUpdated),
					resource.TestCheckResourceAttr("numspot_vm.test", "performance", performanceUpdated),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_vm.test", "tags.0.value", tagValueUpdated),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testVmConfig_UpdateNoReplace(sourceImageId string, vmType string, vmInitiatedShutdownBehavior string, tagKey string, tagValue string, performance string) string {
	return fmt.Sprintf(`
resource "numspot_vm" "test" {
  image_id                       = %[1]q
  type                           = %[2]q
  vm_initiated_shutdown_behavior = %[3]q
  performance                    = %[6]q

  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
}
`, sourceImageId, vmType, vmInitiatedShutdownBehavior, tagKey, tagValue, performance)
}

func TestAccVmResource_Update_WithReplace(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	subregionName := "cloudgouv-eu-west-1a"
	subregionNameUpdated := "cloudgouv-eu-west-1b"
	var vm_id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_UpdateWithReplace(sourceImageId, vmType, subregionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						vm_id = v
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vm.test", "type", vmType),
					resource.TestCheckResourceAttr("numspot_vm.test", "placement.availability_zone_name", subregionName),
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
				Config: testVmConfig_UpdateWithReplace(sourceImageId, vmType, subregionNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						if vm_id == v {
							return errors.New("Id should be different after Update with replace")
						}
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vm.test", "type", vmType),
					resource.TestCheckResourceAttr("numspot_vm.test", "placement.availability_zone_name", subregionNameUpdated),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testVmConfig_UpdateWithReplace(sourceImageId, vmType, subregionName string) string {
	return fmt.Sprintf(`
resource "numspot_vm" "test" {
  image_id = %[1]q
  type     = %[2]q
  placement = {
    tenancy                = "default"
    availability_zone_name = %[3]q
  }
}
`, sourceImageId, vmType, subregionName)
}
