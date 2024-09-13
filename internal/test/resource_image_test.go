package test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataImage struct {
	name,
	sourceImageId,
	isPublic,
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_image.test' has input data values
func getFieldMatchChecksImage(data StepDataImage) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_image.test", "name", data.name),
		resource.TestCheckResourceAttr("numspot_image.test", "access.is_public", data.isPublic),
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
		acctest.TestCheckTypeSetElemNestedAttrsWithPair("numspot_image.test", "block_device_mappings.*", map[string]string{
			"bsu.snapshot_id": fmt.Sprintf(acctest.PAIR_PREFIX+"numspot_snapshot.test%[1]s.id", dependenciesSuffix),
		}),
	}
}

func TestAccImageResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place

	// resource fields that cannot be updated in-place (requires replace)

	name := "terraform-image-test"
	nameUpdated := "terraform-image-test-Updated"
	sourceImageId := "ami-0b7df82c"
	sourceImageIdUpdated := "ami-0987a84b"

	/////////////////////////////////////////////////////////////////////////////////////
	tagKey := "name"
	tagValue := "Terraform-Test-Image"
	tagValueUpdated := tagValue + "-Updated"
	isPublic := "true"
	isPublicUpdated := "false"

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataImage{
		name:          name,
		sourceImageId: sourceImageId,
		tagKey:        tagKey,
		tagValue:      tagValue,
		isPublic:      isPublic,
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
		tagValue:      tagValue,
		isPublic:      isPublic,
	}
	replaceChecks := append(
		getFieldMatchChecksImage(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.NotEqual(t, v, resourceId)
			resourceId = v
			return nil
		}),
	)
	// The plan that should trigger Update behavior (based on replacePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataImage{
		name:          nameUpdated,
		sourceImageId: sourceImageIdUpdated,
		tagKey:        tagKey,
		tagValue:      tagValueUpdated,
		isPublic:      isPublicUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksImage(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_image.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			resourceId = v
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
			},
			// Update testing Without Replace (create image from Image)
			{
				Config: testImageConfig_FromImage(updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(
					append(
						slices.Concat(updateChecks),
						resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", replacePlanValues.sourceImageId),
					)...,
				),
			},
			// Update testing With Replace (create image from Vm)
			{
				Config: testImageConfig_FromVm(acctest.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksImage_FromVm(acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing Without Replace (create image from Vm)
			{
				Config: testImageConfig_FromVm(acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksImage_FromVm(acctest.BASE_SUFFIX),
				)...),
			},

			// Update testing With Replace (create image from Snapshot)
			{
				Config: testImageConfig_FromSnapshot(acctest.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksImage_FromSnapshot(acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing Without Replace (create image from Snapshot)
			{
				Config: testImageConfig_FromSnapshot(acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksImage_FromSnapshot(acctest.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testImageConfig_FromVm(acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testImageConfig_FromVm(acctest.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksImage_FromVm(acctest.NEW_SUFFIX),
				)...),
			},
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testImageConfig_FromSnapshot(acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testImageConfig_FromSnapshot(acctest.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksImage_FromSnapshot(acctest.NEW_SUFFIX),
				)...),
			},

			// <== If resource has optional dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testImageConfig_FromVm(acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Deletion of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testImageConfig_FromImage(replacePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(replaceChecks...),
			},

			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testImageConfig_FromSnapshot(acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Deletion of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testImageConfig_FromImage(replacePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(replaceChecks...),
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
  access = {
    is_public = %[5]s
  }
  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, data.name, data.sourceImageId, data.tagKey, data.tagValue, data.isPublic)
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
  access = {
    is_public = %[6]s
  }
  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
}`, subresourceSuffix, data.name, data.sourceImageId, data.tagKey, data.tagValue, data.isPublic)
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
  access = {
    is_public = %[5]s
  }
  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, subresourceSuffix, data.name, data.tagKey, data.tagValue, data.isPublic)
}
