package volume_test

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataVolume struct {
	volumeType,
	volumeSize,
	tagKey,
	tagValue,
	az,
	deviceName string
}

// Generate checks to validate that resource 'numspot_volume.test' has input data values
func getFieldMatchChecksVolume(data StepDataVolume) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", data.az),
		resource.TestCheckResourceAttr("numspot_volume.test", "type", data.volumeType),
		resource.TestCheckResourceAttr("numspot_volume.test", "size", data.volumeSize),
		resource.TestCheckResourceAttr("numspot_volume.test", "link_vm.device_name", data.deviceName),
		resource.TestCheckResourceAttr("numspot_volume.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_volume.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_volume.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksVolume(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_volume.test", "link_vm.vm_id", "numspot_vm.test"+dependenciesSuffix, "id"),
	}
}

func TestAccVolumeResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "Name"
	tagValue := "terraform-vm"
	tagValueUpdated := tagValue + "-Updated"
	volumeType := "standard"
	volumeTypeUpdated := "gp2"
	volumeSize := "11"
	volumeSizeUpdated := "22"
	volumeAZ := "cloudgouv-eu-west-1a"
	volumeAZUpdated := "cloudgouv-eu-west-1a"
	deviceName := "/dev/sdb"
	deviceNameUpdated := "/dev/sdc"
	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataVolume{
		volumeType: volumeType,
		volumeSize: volumeSize,
		az:         volumeAZ,
		tagKey:     tagKey,
		tagValue:   tagValue,
		deviceName: deviceName,
	}
	createChecks := append(
		getFieldMatchChecksVolume(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_volume.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataVolume{
		volumeType: volumeTypeUpdated,
		volumeSize: volumeSizeUpdated,
		az:         volumeAZUpdated,
		tagKey:     tagKey,
		tagValue:   tagValueUpdated,
		deviceName: deviceNameUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksVolume(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_volume.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)
	/////////////////////////////////////////////////////////////////////////////////////

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // Create testing
				Config: testVolumeConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksVolume(provider.BASE_SUFFIX),
				)...),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_volume.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testVolumeConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVolume(provider.BASE_SUFFIX),
				)...),
				ExpectNonEmptyPlan: true,
			},
			// Update testing With Replace of dependency resource and without Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testVolumeConfig(provider.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVolume(provider.NEW_SUFFIX),
				)...),
				ExpectNonEmptyPlan: true,
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testVolumeConfig_DeletedDependencies(updatePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(updateChecks...),
			},
		},
	})
}

func testVolumeConfig(subresourceSuffix string, data StepDataVolume) string {
	volumeSize, _ := strconv.Atoi(data.volumeSize)
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test%[7]s" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_volume" "test" {
  type                   = %[1]q
  size                   = %[2]d
  availability_zone_name = %[3]q
  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
  link_vm = {
    vm_id       = numspot_vm.test%[7]s.id
    device_name = %[6]q
  }
}`, data.volumeType, volumeSize, data.az, data.tagKey, data.tagValue, data.deviceName, subresourceSuffix)
}

func testVolumeConfig_DeletedDependencies(data StepDataVolume) string {
	volumeSize, _ := strconv.Atoi(data.volumeSize)
	return fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = %[1]q
  size                   = %[2]d
  availability_zone_name = %[3]q
  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
}`, data.volumeType, volumeSize, data.az, data.tagKey, data.tagValue)
}
