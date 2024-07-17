//go:build acc

package provider

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataSpace struct {
	spaceName,
	spaceDescription string
}

// Generate checks to validate that resource 'numspot_space.test' has input data values
func getFieldMatchChecksSpace(data StepDataSpace) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_space.test", "name", data.spaceName),
		resource.TestCheckResourceAttr("numspot_space.test", "description", data.spaceDescription),
	}
}

func TestAccSpaceResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	var resourceId string
	organisationID := "67d97ad4-3005-48dc-a392-60a97ab5097c"

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	// None

	// resource fields that cannot be updated in-place (requires replace)
	spaceName := "quiet space"
	spaceDescription := "A quiet space"
	updatedSpaceName := "best space"
	updatedSpaceDescription := "A quieter space"

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataSpace{
		spaceName:        spaceName,
		spaceDescription: spaceDescription,
	}
	createChecks := append(
		getFieldMatchChecksSpace(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_space.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataSpace{
		spaceName:        updatedSpaceName, // Update values for non-updatable fields
		spaceDescription: updatedSpaceDescription,
	}
	replaceChecks := append(
		getFieldMatchChecksSpace(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_space.test", "id", func(v string) error {
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
				Config: testSpaceConfig(organisationID, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_space.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},

			// Update testing With Replace (if needed)
			{
				Config: testSpaceConfig(organisationID, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
				)...),
			},
		},
	})
}

func testSpaceConfig(organisationID string, data StepDataSpace) string {
	return fmt.Sprintf(`
resource "numspot_space" "test" {
  organisation_id = %[1]q
  name            = %[2]q
  description     = %[3]q
}`, organisationID, data.spaceName, data.spaceDescription)
}
