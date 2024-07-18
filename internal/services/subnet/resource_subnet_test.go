//go:build acc

package subnet_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataSubnet struct {
	tagKey,
	tagValue,
	mapPublicIpOnLaunch,
	netIpRange,
	subnetIpRange string
}

// Generate checks to validate that resource 'numspot_subnet.test' has input data values
func getFieldMatchChecksSubnet(data StepDataSubnet) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_subnet.test", "map_public_ip_on_launch", data.mapPublicIpOnLaunch),
		resource.TestCheckResourceAttr("numspot_subnet.test", "net_ip_range", data.netIpRange),
		resource.TestCheckResourceAttr("numspot_subnet.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_subnet.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_subnet.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksSubnet(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_subnet.test", "vpc_id", "numspot_vpc.test"+dependenciesPrefix, "id"),
	}
}

func TestAccSubnetResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories
	subnetIpRange := "10.101.1.0/24"

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagName := "name"
	tagValue := "subnet"
	tagValueUpdated := "subnet updated"

	mapPublicIpOnLaunch := "true"
	mapPublicIpOnLaunchUpdated := "false"

	// resource fields that cannot be updated in-place (requires replace)
	netIpRange := "10.101.0.0/16"
	netIpRangeUpdated := "10.101.2.0/24"

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataSubnet{
		tagKey:              tagName,
		tagValue:            tagValue,
		mapPublicIpOnLaunch: mapPublicIpOnLaunch,
		netIpRange:          netIpRange,
		subnetIpRange:       subnetIpRange,
	}
	createChecks := append(
		getFieldMatchChecksSubnet(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataSubnet{
		tagKey:              tagName,
		tagValue:            tagValueUpdated,
		mapPublicIpOnLaunch: mapPublicIpOnLaunchUpdated,
		netIpRange:          netIpRange,
		subnetIpRange:       subnetIpRange,
	}
	updateChecks := append(
		getFieldMatchChecksSubnet(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataSubnet{
		tagKey:              tagName,
		tagValue:            tagValue,
		mapPublicIpOnLaunch: mapPublicIpOnLaunch,
		netIpRange:          netIpRangeUpdated,
		subnetIpRange:       subnetIpRange,
	}
	replaceChecks := append(
		getFieldMatchChecksSubnet(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_subnet.test", "id", func(v string) error {
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
				Config: testSubnetConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksSubnet(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_subnet.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testSubnetConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksSubnet(provider.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testSubnetConfig(provider.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksSubnet(provider.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testSubnetConfig(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testSubnetConfig(provider.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksSubnet(provider.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func testSubnetConfig(subresourceSuffix string, data StepDataSubnet) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test%[1]s" {
  ip_range = %[2]q
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.test%[1]s.id
  ip_range                = %[3]q
  map_public_ip_on_launch = %[4]s
  tags = [
    {
      key   = %[5]q
      value = %[6]q
    }
  ]

}`, subresourceSuffix, data.subnetIpRange, data.netIpRange, data.mapPublicIpOnLaunch, data.tagKey, data.tagValue)
}
