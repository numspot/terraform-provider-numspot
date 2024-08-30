//go:build acc

package publicip_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataPublicIp struct {
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_public_ip.test' has input data values
func getFieldMatchChecksPublicIp(data StepDataPublicIp) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_public_ip.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_public_ip.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksPublicIp(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_public_ip.test", "vm_id", "numspot_vm.test"+dependenciesSuffix, "id"),
	}
}

func TestAccPublicIpResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place

	tagKey := "name"
	tagValue := "Terraform-Test-Public-Ip"
	tagValueUpdated := "Terraform-Test-Public-Ip-Updated"

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataPublicIp{
		tagKey:   tagKey,
		tagValue: tagValue,
	}
	createChecks := append(
		getFieldMatchChecksPublicIp(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_public_ip.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataPublicIp{
		tagKey:   tagKey,
		tagValue: tagValueUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksPublicIp(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_public_ip.test", "id", func(v string) error {
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
				Config: testPublicIpConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksPublicIp(provider.BASE_SUFFIX),
				)...),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_public_ip.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testPublicIpConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksPublicIp(provider.BASE_SUFFIX),
				)...),
				ExpectNonEmptyPlan: true,
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

func testPublicIpConfig(subresourceSuffix string, data StepDataPublicIp) string {
	return fmt.Sprintf(`
resource "numspot_image" "test" {
  name               = "terraform-generated-image-for-public-ip-acctest"
  source_image_id    = "ami-0b7df82c"
  source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test%[1]s" {
  image_id  = numspot_image.test.id
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_public_ip" "test" {
  vm_id      = numspot_vm.test%[1]s.id
  depends_on = [numspot_internet_gateway.test]
  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, subresourceSuffix, data.tagKey, data.tagValue)
}

// <== If resource has optional dependencies ==>
func testPublicIpConfig_DeletedDependencies(data StepDataPublicIp) string {
	return fmt.Sprintf(`
resource "numspot_public_ip" "test" {
  tags = [
    {
      key   = %[1]q
      value = %[2]q
    }
  ]
}`, data.tagKey, data.tagValue)
}
