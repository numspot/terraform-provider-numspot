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
type StepDataVpcPeering struct {
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_vpc_peering.test' has input data values
func getFieldMatchChecksVpcPeering(data StepDataVpcPeering) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc_peering.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_vpc_peering.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksVpcPeering(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_vpc_peering.test", "accepter_vpc_id", "numspot_vpc.accepter"+dependenciesPrefix, "id"),
		resource.TestCheckResourceAttrPair("numspot_vpc_peering.test", "source_vpc_id", "numspot_vpc.source"+dependenciesPrefix, "id"),
	}
}

func TestAccVpcPeeringResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "name"
	tagValue := "Terraform-Test-Vpc-Peering"
	tagValueUpdated := "Terraform-Test-Vpc-Peering-2"

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataVpcPeering{
		tagKey:   tagKey,
		tagValue: tagValue,
	}
	createChecks := append(
		getFieldMatchChecksVpcPeering(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataVpcPeering{
		tagKey:   tagKey,
		tagValue: tagValueUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksVpcPeering(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataVpcPeering{ // Note : replace when dependencies are changed, no other attributes induces replace
		tagKey:   tagKey,
		tagValue: tagValue,
	}
	replaceChecks := append(
		getFieldMatchChecksVpcPeering(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "id", func(v string) error {
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
				Config: testVpcPeeringConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksVpcPeering(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing Without Replace (if needed)
			{
				Config: testVpcPeeringConfig(utils_acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVpcPeering(utils_acctest.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testVpcPeeringConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testVpcPeeringConfig(utils_acctest.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksVpcPeering(utils_acctest.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func testVpcPeeringConfig(subresourceSuffix string, data StepDataVpcPeering) string {
	return fmt.Sprintf(`

resource "numspot_vpc" "accepter%[1]s" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "source%[1]s" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter%[1]s.id
  source_vpc_id   = numspot_vpc.source%[1]s.id

  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, subresourceSuffix, data.tagKey, data.tagValue)
}
