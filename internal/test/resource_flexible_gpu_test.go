package test

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataFlexibleGpu struct {
	modelName,
	generation,
	az string
}

// Generate checks to validate that resource 'numspot_flexible_gpu.test' has input data values
func getFieldMatchChecksFlexibleGpu(data StepDataFlexibleGpu) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", data.modelName),      // Check value for all resource attributes
		resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", data.generation),     // Check value for all resource attributes
		resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", data.az), // Check value for all resource attributes
	}
}

// Generate checks to validate that resource 'numspot_flexible_gpu.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksFlexibleGpu(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_flexible_gpu.test", "vm_id", "numspot_vm.test"+dependenciesSuffix, "id"),
	}
}

func TestAccFlexibleGpuResource(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skipf("skipping %s test in CI", t.Name())
	}
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place

	modelName := "nvidia-a100-80"
	generation := "v6"
	az := "cloudgouv-eu-west-1a"

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataFlexibleGpu{
		modelName:  modelName,
		generation: generation,
		az:         az,
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
		modelName:  modelName,
		generation: generation,
		az:         az,
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
				Config: testFlexibleGpuConfig(acctest.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksFlexibleGpu(acctest.BASE_SUFFIX),
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
				Config: testFlexibleGpuConfig(acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksFlexibleGpu(acctest.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testFlexibleGpuConfig(acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			{
				Config: testFlexibleGpuConfig(acctest.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksFlexibleGpu(acctest.NEW_SUFFIX),
				)...),
			},

			// <== If resource has optional dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testFlexibleGpuConfig(acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
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

resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.net.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = %[4]q
}

resource "numspot_vm" "test%[1]s" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_flexible_gpu" "test" {
  model_name             = %[2]q
  generation             = %[3]q
  availability_zone_name = %[4]q
  vm_id                  = numspot_vm.test%[1]s.id
}`, subresourceSuffix, data.modelName, data.generation, data.az)
}

// <== If resource has optional dependencies ==>
func testFlexibleGpuConfig_DeletedDependencies(data StepDataFlexibleGpu) string {
	return fmt.Sprintf(`
resource "numspot_flexible_gpu" "test" {
  model_name             = %[1]q
  generation             = %[2]q
  availability_zone_name = %[3]q
}`, data.modelName, data.generation, data.az)
}
