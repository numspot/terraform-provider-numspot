//go:build acc

package nic_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataNic struct {
	tagKey,
	tagValue,
	description string
}

// Generate checks to validate that resource 'numspot_nic.test' has input data values
func getFieldMatchChecksNic(data StepDataNic) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_nic.test", "description", data.description),
		resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_nic.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_nic.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksNic(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_nic.test", "subnet_id", "numspot_subnet.test"+dependenciesPrefix, "id"),
		resource.TestCheckTypeSetElemAttrPair("numspot_nic.test", "security_group_ids.*", "numspot_numspot_security_group.test"+dependenciesPrefix, "id"),
	}
}

func TestAccNicResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "name"
	tagValue := "Terraform-Test-Volume"
	tagValueUpdated := "Terraform-Test-Volume-Update"

	description := "The nic"
	descriptionUpdated := "The better nic"
	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataNic{
		tagKey:      tagKey,
		tagValue:    tagValue,
		description: description,
	}
	createChecks := append(
		getFieldMatchChecksNic(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataNic{
		tagKey:      tagKey,
		tagValue:    tagValueUpdated,
		description: descriptionUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksNic(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataNic{ // Replace is triggered by a dependency change (subnet)
		tagKey:      tagKey,
		tagValue:    tagValue,
		description: description,
	}
	replaceChecks := append(
		getFieldMatchChecksNic(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
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
				Config: testNicConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksNic(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testNicConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksNic(provider.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testNicConfig(provider.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksNic(provider.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testNicConfig(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testNicConfig(provider.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksNic(provider.NEW_SUFFIX),
				)...),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testNicConfig(provider.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksNic(provider.NEW_SUFFIX),
				)...),
			},

			// <== If resource has optional dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testNicConfig(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testNicConfig_DeletedDependencies(updatePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(updateChecks...),
			},
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testNicConfig(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Deletion of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testNicConfig_DeletedDependencies(replacePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(replaceChecks...),
			},
		},
	})
}

func testNicConfig(subresourceSuffix string, data StepDataNic) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test%[1]s" {
  vpc_id   = numspot_vpc.test.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test%[1]s" {
  vpc_id      = numspot_vpc.test.id
  name        = "security_group"
  description = "numspot_security_group description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_nic" "test" {
  subnet_id          = numspot_subnet.test%[1]s.id
  description        = %[2]q
  security_group_ids = [numspot_security_group.test%[1]s.id]
  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, subresourceSuffix, data.description, data.tagKey, data.tagValue)
}

// <== If resource has optional dependencies ==>
func testNicConfig_DeletedDependencies(data StepDataNic) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id   = numspot_subnet.test.id
  description = %[1]q
  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, data.description, data.tagKey, data.tagValue)
}
