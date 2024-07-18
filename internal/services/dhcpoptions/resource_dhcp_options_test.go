//go:build acc

package dhcpoptions_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataDhcpOptions struct {
	domain,
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_dhcp_options.test' has input data values
func getFieldMatchChecksDhcpOptions(data StepDataDhcpOptions) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_dhcp_options.test", "domain", data.domain), // Check value for all resource attributes
		resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_dhcp_options.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksDhcpOptions(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{}
}

func TestAccDhcpOptionsResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "name"
	tagName := "Terraform Provider DHCP Options"
	updatedTagName := "Terraform Provider DHCP Options - 2"

	// resource fields that cannot be updated in-place (requires replace)
	domainName := "foo.bar"
	updatedDomainName := "bar.foo"

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataDhcpOptions{
		domain:   domainName,
		tagKey:   tagKey,
		tagValue: tagName,
	}
	createChecks := append(
		getFieldMatchChecksDhcpOptions(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_dhcp_options.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataDhcpOptions{
		domain:   updatedDomainName,
		tagKey:   tagKey,
		tagValue: updatedTagName,
	}
	updateChecks := append(
		getFieldMatchChecksDhcpOptions(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_dhcp_options.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataDhcpOptions{
		domain:   domainName,
		tagKey:   tagKey,
		tagValue: tagName,
	}
	replaceChecks := append(
		getFieldMatchChecksDhcpOptions(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_dhcp_options.test", "id", func(v string) error {
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
				Config: testDhcpOptionsConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksDhcpOptions(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_dhcp_options.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testDhcpOptionsConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksDhcpOptions(provider.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testDhcpOptionsConfig(provider.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksDhcpOptions(provider.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testDhcpOptionsConfig(provider.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testDhcpOptionsConfig(provider.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksDhcpOptions(provider.NEW_SUFFIX),
				)...),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testDhcpOptionsConfig(provider.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksDhcpOptions(provider.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func testDhcpOptionsConfig(subresourceSuffix string, data StepDataDhcpOptions) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "test" {
  domain_name = %[1]q
  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, data.domain, data.tagKey, data.tagValue)
}
