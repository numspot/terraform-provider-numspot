//go:build acc

package snapshot_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataSnapshot struct {
	tagKey,
	tagValue,
	description string
}

// Generate checks to validate that resource 'numspot_snapshot.test' has input data values
func getFieldMatchChecksSnapshot(data StepDataSnapshot) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_snapshot.test", "description", data.description),
		resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_snapshot.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksSnapshot_FromSnapshot(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_snapshot.test", "source_snapshot_id", "numspot_snapshot.snapshot"+dependenciesPrefix, "id"),
	}
}

func getDependencyChecksSnapshot_FromVolume(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_snapshot.test", "volume_id", "numspot_volume.test"+dependenciesPrefix, "id"),
	}
}

func TestAccSnapshotResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	tagKey := "name"

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagValue := "Snapshot-Terraform-Test"
	tagValueUpdated := "Snapshot-Terraform-Test-Updated"

	// resource fields that cannot be updated in-place (requires replace)
	description := "A beautiful snapshot"
	descriptionUpdated := "An even more beautiful snapshot"

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataSnapshot{
		tagKey:      tagKey,
		tagValue:    tagValue,
		description: description,
	}
	createChecks := append(
		getFieldMatchChecksSnapshot(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataSnapshot{
		tagKey:      tagKey,
		tagValue:    tagValueUpdated,
		description: description,
	}
	updateChecks := append(
		getFieldMatchChecksSnapshot(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataSnapshot{
		tagKey:      tagKey,
		tagValue:    tagValue,
		description: descriptionUpdated,
	}
	replaceChecks := append(
		getFieldMatchChecksSnapshot(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.NotEqual(t, v, resourceId)
			return nil
		}),
	)
	/////////////////////////////////////////////////////////////////////////////////////

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			//////// Test snapshot created from Volume
			{ // Create testing
				Config: testSnapshotConfig_FromVolume(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksSnapshot_FromVolume(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testSnapshotConfig_FromVolume(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksSnapshot_FromVolume(provider.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testSnapshotConfig_FromVolume(provider.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksSnapshot_FromVolume(provider.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testSnapshotConfig_FromVolume(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testSnapshotConfig_FromVolume(provider.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksSnapshot_FromVolume(provider.NEW_SUFFIX),
				)...),
			},

			//////// Test snapshot created from another Snapshot
			{ // Create testing
				Config: testSnapshotConfig_FromSnapshot(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksSnapshot_FromSnapshot(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_region_name", "source_snapshot_id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testSnapshotConfig_FromSnapshot(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksSnapshot_FromSnapshot(provider.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testSnapshotConfig_FromSnapshot(provider.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksSnapshot_FromSnapshot(provider.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testSnapshotConfig_FromSnapshot(provider.BASE_SUFFIX, basePlanValues),
			},

			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testSnapshotConfig_FromSnapshot(provider.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksSnapshot_FromSnapshot(provider.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func testSnapshotConfig_FromVolume(subresourceSuffix string, data StepDataSnapshot) string {
	return fmt.Sprintf(`

resource "numspot_volume" "test%[1]s" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test%[1]s.id
  description = %[2]q
  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, subresourceSuffix, data.description, data.tagKey, data.tagValue)
}

func testSnapshotConfig_FromSnapshot(subresourceSuffix string, data StepDataSnapshot) string {
	return fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "snapshot%[1]s" {
  volume_id = numspot_volume.test.id
}

resource "numspot_snapshot" "test" {
  source_snapshot_id = numspot_snapshot.snapshot%[1]s.id
  source_region_name = "cloudgouv-eu-west-1"
  description        = %[2]q
  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, subresourceSuffix, data.description, data.tagKey, data.tagValue)
}
