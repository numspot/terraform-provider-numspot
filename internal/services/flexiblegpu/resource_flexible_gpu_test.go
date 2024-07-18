//go:build acc

package flexiblegpu_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataFlexibleGpu struct {
	modelName,
	generation,
	az,
	deleteOnVMDeletion string
}

// Generate checks to validate that resource 'numspot_flexible_gpu.test' has input data values
func getFieldMatchChecksFlexibleGpu(data StepDataFlexibleGpu) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", data.modelName),                     // Check value for all resource attributes
		resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", data.generation),                    // Check value for all resource attributes
		resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", data.az),                // Check value for all resource attributes
		resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "delete_on_vm_deletion", data.deleteOnVMDeletion), // Check value for all resource attributes
	}
}

// Generate checks to validate that resource 'numspot_flexible_gpu.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksFlexibleGpu(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_flexible_gpu.test", "vm_id", "numspot_vm.test"+dependenciesPrefix, "id"),
	}
}

func TestAccFlexibleGpuResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place

	modelName := "nvidia-a100-80"
	generation := "v6"
	az := "cloudgouv-eu-west-1a"
	deleteOnVMDeletion := "false"
	deleteOnVMDeletionUpdated := "true"

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataFlexibleGpu{
		modelName:          modelName,
		generation:         generation,
		az:                 az,
		deleteOnVMDeletion: deleteOnVMDeletion,
	}
	createChecks := append(
		getFieldMatchChecksFlexibleGpu(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataFlexibleGpu{
		modelName:          modelName,
		generation:         generation,
		az:                 az,
		deleteOnVMDeletion: deleteOnVMDeletionUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksFlexibleGpu(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
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
				Config: testFlexibleGpuConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksFlexibleGpu(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_flexible_gpu.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testFlexibleGpuConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksFlexibleGpu(provider.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testFlexibleGpuConfig(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testFlexibleGpuConfig(provider.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksFlexibleGpu(provider.NEW_SUFFIX),
				)...),
			},

			// <== If resource has optional dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testFlexibleGpuConfig(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testFlexibleGpuConfig_DeletedDependencies(updatePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(updateChecks...),
			},
		},
	})
}

func testFlexibleGpuConfig(subresourceSuffix string, data StepDataFlexibleGpu) string {
	return fmt.Sprintf(`


// <== If resource has dependencies ==> 
resource "numspot_vm" "test%[1]s" {
  image_id = "ami-0b7df82c"
  type     = "ns-mus6-2c16r"
}

resource "numspot_flexible_gpu" "test" {
  model_name             = %[2]q
  generation             = %[3]q
  availability_zone_name = %[4]q
  delete_on_vm_deletion  = %[5]q
  vm_id                  = numspot_vm.test%[1]s.id
}`, subresourceSuffix, data.modelName, data.generation, data.az, data.deleteOnVMDeletion)
}

// <== If resource has optional dependencies ==>
func testFlexibleGpuConfig_DeletedDependencies(data StepDataFlexibleGpu) string {
	return fmt.Sprintf(`
resource "numspot_flexible_gpu" "test" {
  model_name             = %[1]q
  generation             = %[2]q
  availability_zone_name = %[3]q
  delete_on_vm_deletion  = %[4]q
}`, data.modelName, data.generation, data.az, data.deleteOnVMDeletion)
}
