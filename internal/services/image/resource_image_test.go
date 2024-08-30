//go:build acc

package image_test

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataImage struct {
	name,
	sourceImageId,
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_image.test' has input data values
func getFieldMatchChecksImage(data StepDataImage) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_image.test", "name", data.name),
		resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_image.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_image.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksImage_FromVm(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_image.test", "vm_id", "numspot_vm.test"+dependenciesSuffix, "id"),
	}
}

func getDependencyChecksImage_FromSnapshot(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		provider.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test", "block_device_mappings.*", map[string]string{
			"bsu.snapshot_id": fmt.Sprintf(provider.PAIR_PREFIX+"numspot_snapshot.test%[1]s.id", dependenciesSuffix),
		}),
	}
}

func TestAccImageResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place

	// resource fields that cannot be updated in-place (requires replace)

	randint := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("terraform-image-test-%d", randint)
	nameUpdated := fmt.Sprintf("terraform-image-test-Updated-%d", randint)
	sourceImageId := "ami-0b7df82c"
	sourceImageIdUpdated := "ami-0987a84b"

	/////////////////////////////////////////////////////////////////////////////////////
	tagKey := "name"
	tagValue := "Terraform-Test-Image"
	tagValueUpdated := tagValue + "-Updated"

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataImage{
		name:          name,
		sourceImageId: sourceImageId,
		tagKey:        tagKey,
		tagValue:      tagValue,
	}
	createChecks := append(
		getFieldMatchChecksImage(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataImage{
		name:          nameUpdated,
		sourceImageId: sourceImageIdUpdated,
		tagKey:        tagKey,
		tagValue:      tagValueUpdated,
	}
	replaceChecks := append(
		getFieldMatchChecksImage(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
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
				Config: testImageConfig_FromImage(basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(
					append(
						slices.Concat(createChecks),
						resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", basePlanValues.sourceImageId),
					)...,
				),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_image_id", "source_region_name"},
			},
			// Update testing With Replace (create image from Image)
			{
				Config: testImageConfig_FromImage(replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(
					append(
						slices.Concat(replaceChecks),
						resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", replacePlanValues.sourceImageId),
					)...,
				),
				ExpectNonEmptyPlan: true,
			},
			// Update testing With Replace (create image from Vm)
			{
				Config: testImageConfig_FromVm(provider.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksImage_FromVm(provider.BASE_SUFFIX),
				)...),
				ExpectNonEmptyPlan: true,
			},
			// Update testing With Replace (create image from Snapshot)
			{
				Config: testImageConfig_FromSnapshot(provider.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksImage_FromSnapshot(provider.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config:             testImageConfig_FromVm(provider.BASE_SUFFIX, basePlanValues),
				ExpectNonEmptyPlan: true,
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testImageConfig_FromVm(provider.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksImage_FromVm(provider.NEW_SUFFIX),
				)...),
				ExpectNonEmptyPlan: true,
			},
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testImageConfig_FromSnapshot(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testImageConfig_FromSnapshot(provider.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksImage_FromSnapshot(provider.NEW_SUFFIX),
				)...),
			},

			// <== If resource has optional dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config:             testImageConfig_FromVm(provider.BASE_SUFFIX, basePlanValues),
				ExpectNonEmptyPlan: true,
			},
			// Update testing With Deletion of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config:             testImageConfig_FromImage(replacePlanValues),
				Check:              resource.ComposeAggregateTestCheckFunc(replaceChecks...),
				ExpectNonEmptyPlan: true,
			},

			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testImageConfig_FromSnapshot(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Deletion of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config:             testImageConfig_FromImage(replacePlanValues),
				Check:              resource.ComposeAggregateTestCheckFunc(replaceChecks...),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testImageConfig_FromImage(data StepDataImage) string {
	return fmt.Sprintf(`
resource "numspot_image" "test" {
  name               = %[1]q
  source_image_id    = %[2]q
  source_region_name = "cloudgouv-eu-west-1"
  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, data.name, data.sourceImageId, data.tagKey, data.tagValue)
}

// TODO test availability zone with placement VM directly instead of nested subnet
func testImageConfig_FromVm(subresourceSuffix string, data StepDataImage) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}
resource "numspot_vm" "test%[1]s" {
  image_id  = %[3]q
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}
resource "numspot_image" "test" {
  name  = %[2]q
  vm_id = numspot_vm.test%[1]s.id
  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
}`, subresourceSuffix, data.name, data.sourceImageId, data.tagKey, data.tagValue)
}

func testImageConfig_FromSnapshot(subresourceSuffix string, data StepDataImage) string {
	return fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1b"
}

resource "numspot_snapshot" "test%[1]s" {
  volume_id   = numspot_volume.test.id
  description = "a numspot snapshot description"
}

resource "numspot_image" "test" {
  name             = %[2]q
  root_device_name = "/dev/sda1"
  block_device_mappings = [
    {
      device_name = "/dev/sda1"
      bsu = {
        snapshot_id           = numspot_snapshot.test%[1]s.id
        volume_size           = 120
        volume_type           = "io1"
        iops                  = 150
        delete_on_vm_deletion = true
      }
    }
  ]
  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, subresourceSuffix, data.name, data.tagKey, data.tagValue)
}
