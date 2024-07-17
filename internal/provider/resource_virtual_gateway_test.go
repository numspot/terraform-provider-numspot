//go:build acc

package provider

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataVirtualGateway struct {
	connectionType,
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_virtual_gateway.test' has input data values
func getFieldMatchChecksVirtualGateway(data StepDataVirtualGateway) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", data.connectionType), // Check value for all resource attributes
		resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_virtual_gateway.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_virtual_gateway.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksVirtualGateway(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", "numspot_vpc.test"+dependenciesPrefix, "id"),
	}
}

func TestAccVirtualGatewayResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	var resourceId string

	connectionType := "ipsec.1"
	tagKey := "name"

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagValue := "Terraform-Test-Volume"
	tagValueUpdate := tagValue + "-Update"

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataVirtualGateway{
		connectionType: connectionType,
		tagKey:         tagKey,
		tagValue:       tagValue,
	}
	createChecks := append(
		getFieldMatchChecksVirtualGateway(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataVirtualGateway{
		connectionType: connectionType,
		tagKey:         tagKey,
		tagValue:       tagValueUpdate,
	}
	updateChecks := append(
		getFieldMatchChecksVirtualGateway(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
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
				Config: testVirtualGatewayConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksVirtualGateway(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_virtual_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testVirtualGatewayConfig(utils_acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVirtualGateway(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testVirtualGatewayConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testVirtualGatewayConfig(utils_acctest.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVirtualGateway(utils_acctest.NEW_SUFFIX),
				)...),
			},

			// <== If resource has optional dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testVirtualGatewayConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testVirtualGatewayConfig_DeletedDependencies(updatePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(updateChecks...),
			},
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testVirtualGatewayConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
			},
		},
	})
}

func testVirtualGatewayConfig(subresourceSuffix string, data StepDataVirtualGateway) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test%[1]s" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = %[2]q
  vpc_id          = numspot_vpc.test%[1]s.id
  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, subresourceSuffix, data.connectionType, data.tagKey, data.tagValue)
}

// <== If resource has optional dependencies ==>
func testVirtualGatewayConfig_DeletedDependencies(data StepDataVirtualGateway) string {
	return fmt.Sprintf(`
resource "numspot_virtual_gateway" "test" {
  connection_type = %[1]q
  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, data.connectionType, data.tagKey, data.tagValue)
}
