//go:build acc

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
	az string
}

// Generate checks to validate that resource 'numspot_volume.test' has input data values
func getFieldMatchChecksVolume(data StepDataVolume) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", data.az),
		resource.TestCheckResourceAttr("numspot_volume.test", "type", data.volumeType),
		resource.TestCheckResourceAttr("numspot_volume.test", "size", data.volumeSize),
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
	return []resource.TestCheckFunc{}
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
				Config: testVolumeConfig(basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksVolume(provider.BASE_SUFFIX),
				)...),
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
				Config: testVolumeConfig(updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVolume(provider.BASE_SUFFIX),
				)...),
			},
		},
	})
}

func testVolumeConfig(data StepDataVolume) string {
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
