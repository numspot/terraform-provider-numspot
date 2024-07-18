//go:build acc

package routetable_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataRouteTable struct {
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_route_table.test' has input data values
func getFieldMatchChecksRouteTable(data StepDataRouteTable) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_route_table.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_route_table.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_route_table.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksRouteTable(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_route_table.test", "subnet_id", "numspot_subnet.test"+dependenciesPrefix, "id"),
		resource.TestCheckResourceAttrPair("numspot_route_table.test", "vpc_id", "numspot_vpc.test"+dependenciesPrefix, "id"),
		provider.TestCheckTypeSetElemNestedAttrsWithPair("numspot_route_table.test", "routes.*", map[string]string{ // If field is a list of objects (containing id and/or other fields)
			"gateway_id": fmt.Sprintf(provider.PAIR_PREFIX+"numspot_internet_gateway.test%[1]s.id", dependenciesPrefix),
		}),
	}
}

func TestAccRouteTableResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "name"
	tagValue := "Terraform-Test-Volume"
	tagValueUpdate := "Terraform-Test-Volume-Update"

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataRouteTable{
		tagKey:   tagKey,
		tagValue: tagValue,
	}
	createChecks := append(
		getFieldMatchChecksRouteTable(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataRouteTable{
		tagKey:   tagKey,
		tagValue: tagValueUpdate,
	}
	updateChecks := append(
		getFieldMatchChecksRouteTable(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
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
				Config: testRouteTableConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksRouteTable(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_route_table.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testRouteTableConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksRouteTable(provider.BASE_SUFFIX),
				)...),
			},
			// <== If resource has required dependencies ==>
			// --> DELETED TEST <-- : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly

			// <== If resource has optional dependencies ==>
			// --> DELETED TEST <-- : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
		},
	})
}

func testRouteTableConfig(subresourceSuffix string, data StepDataRouteTable) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test%[1]s" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test%[1]s" {
  vpc_id   = numspot_vpc.test%[1]s.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_internet_gateway" "test%[1]s" {
  vpc_id = numspot_vpc.test%[1]s.id
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.test%[1]s.id
  subnet_id = numspot_subnet.test%[1]s.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test%[1]s.id
    }
  ]
  tags = [{
    key   = %[2]q
    value = %[3]q
  }]
}`, subresourceSuffix, data.tagKey, data.tagValue)
}

// <== If resource has optional dependencies ==>
func testRouteTableConfig_DeletedDependencies(data StepDataRouteTable) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_route_table" "test" {
  vpc_id = numspot_vpc.test.id
  tags = [{
    key   = %[1]q
    value = %[2]q
  }]
}`, data.tagKey, data.tagValue)
}
