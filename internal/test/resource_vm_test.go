package test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

var (
	vmType        = "ns-eco6-2c8r"
	sourceImageId = "ami-0987a84b"
)

type StepDataVm struct {
	sourceImageId,
	vmType,
	vmInitiatedShutdownBehavior,
	tagKey,
	tagValue,
	subregionName string
}

// Generate checks to validate that resource numspot_vm.test has input data values
func getFieldMatchChecksVm(data StepDataVm) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_vm.test", "image_id", data.sourceImageId),
		resource.TestCheckResourceAttr("numspot_vm.test", "type", data.vmType),
		resource.TestCheckResourceAttr("numspot_vm.test", "vm_initiated_shutdown_behavior", data.vmInitiatedShutdownBehavior),
		resource.TestCheckResourceAttr("numspot_vm.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
		resource.TestCheckResourceAttr("numspot_vm.test", "placement.availability_zone_name", data.subregionName),
	}
}

// Generate checks to validate that resource numspot_vm.test has input data values
func getDependencyChecksVm(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_vm.test", "subnet_id", "numspot_subnet.test"+dependenciesSuffix, "id"),
		resource.TestCheckTypeSetElemAttrPair("numspot_vm.test", "security_group_ids.*", "numspot_security_group.test"+dependenciesSuffix, "id"),
	}
}

func TestAccVmResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "Name"
	tagValue := "terraform-vm"
	tagValueUpdated := tagValue + "-Updated"
	vmTypeUpdated := "ns-cus6-4c8r"
	vmInitiatedShutdownBehavior := "stop"
	// vmInitiatedShutdownBehaviorUpdated := "terminate"

	// resource fields that cannot be updated in-place (requires replace)
	subregionName := "cloudgouv-eu-west-1a"
	//subregionNameUpdated := "cloudgouv-eu-west-1b"
	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataVm{
		sourceImageId:               sourceImageId,
		vmType:                      vmType,
		vmInitiatedShutdownBehavior: vmInitiatedShutdownBehavior,
		tagKey:                      tagKey,
		tagValue:                    tagValue,
		subregionName:               subregionName,
	}
	createChecks := append(
		getFieldMatchChecksVm(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataVm{
		sourceImageId:               sourceImageId,
		vmType:                      vmTypeUpdated,
		vmInitiatedShutdownBehavior: vmInitiatedShutdownBehavior,
		tagKey:                      tagKey,
		tagValue:                    tagValueUpdated,
		subregionName:               subregionName,
	}
	updateChecks := append(
		getFieldMatchChecksVm(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.

	replaceChecks := append(
		getFieldMatchChecksVm(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.NotEqual(t, v, resourceId)
			return nil
		}),
	)
	/////////////////////////////////////////////////////////////////////////////////////

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // Create testing
				Config: testVmConfig(acctest.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksVm(acctest.BASE_SUFFIX),
				)...),
			},
			{ // ImportState testing
				ResourceName:            "numspot_vm.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace
			{
				Config: testVmConfig(acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVm(acctest.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>

			// Update testing With Replace of dependency resource and with Replace of the resource
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testVmConfig(acctest.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksVm(acctest.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func testVmConfig(subresourceSuffix string, data StepDataVm) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test%[1]s" {
  vpc_id                 = numspot_vpc.net.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = %[7]q
}

resource "numspot_security_group" "test%[1]s" {
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
  image_id                       = %[2]q
  type                           = %[3]q
  vm_initiated_shutdown_behavior = %[4]q


  tags = [
    {
      key   = %[5]q
      value = %[6]q
    }
  ]

  placement = {
    tenancy                = "default"
    availability_zone_name = %[7]q
  }

  subnet_id          = numspot_subnet.test%[1]s.id
  security_group_ids = [numspot_security_group.test%[1]s.id]
}
`, subresourceSuffix,
		data.sourceImageId,
		data.vmType,
		data.vmInitiatedShutdownBehavior,
		data.tagKey,
		data.tagValue,
		data.subregionName,
	)
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
				Config: testVmConfig_NetSubnetSGRouteTable(sourceImageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "type", vmType),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "vpc_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("numspot_vm.test", "id"),
				),
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
