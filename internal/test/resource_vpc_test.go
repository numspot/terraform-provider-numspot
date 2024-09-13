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
type StepDataVpc struct {
	ipRange,
	tenancy,
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_vpc.test' has input data values
func getFieldMatchChecksVpc(data StepDataVpc) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", data.ipRange),
		resource.TestCheckResourceAttr("numspot_vpc.test", "tenancy", data.tenancy),
		resource.TestCheckResourceAttr("numspot_vpc.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_vpc.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_vpc.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksVpc(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_vpc.test", "dhcp_options_set_id", "numspot_dhcp_options.test"+dependenciesSuffix, "id"),
	}
}

func TestAccVpcResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "name"
	tagValue := "Terraform Provider VPC"
	tagValueUpdated := "Terraform Provider VPC - 2"

	// resource fields that cannot be updated in-place (requires replace)
	ipRange := "10.101.0.0/16"
	ipRangeUpdated := "10.102.0.0/16"
	tenancy := "default"
	tenancyUpdated := "dedicated"

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataVpc{
		ipRange:  ipRange,
		tenancy:  tenancy,
		tagKey:   tagKey,
		tagValue: tagValue,
	}
	createChecks := append(
		getFieldMatchChecksVpc(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataVpc{
		ipRange:  ipRange,
		tenancy:  tenancy,
		tagKey:   tagKey,
		tagValue: tagValueUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksVpc(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataVpc{
		ipRange:  ipRangeUpdated,
		tenancy:  tenancyUpdated,
		tagKey:   tagKey,
		tagValue: tagValue,
	}
	replaceChecks := append(
		getFieldMatchChecksVpc(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
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
				Config: testVpcConfig(acctest.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksVpc(acctest.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testVpcConfig(acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVpc(acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testVpcConfig(acctest.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksVpc(acctest.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testVpcConfig(acctest.BASE_SUFFIX, basePlanValues),
			},
			// --> DELETED TEST <-- : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly

			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testVpcConfig(acctest.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksVpc(acctest.NEW_SUFFIX),
				)...),
			},

			// <== If resource has optional dependencies ==>
			// --> DELETED TEST <-- : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)

			// --> DELETED TEST <-- : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			// Update testing With Deletion of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly (empty dependency)
		},
	})
}

func testVpcConfig(subresourceSuffix string, data StepDataVpc) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "test%[1]s" {
  domain_name = "domain"
}

resource "numspot_vpc" "test" {
  ip_range            = %[2]q
  dhcp_options_set_id = numspot_dhcp_options.test%[1]s.id
  tenancy             = %[3]q
  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
}`, subresourceSuffix, data.ipRange, data.tenancy, data.tagKey, data.tagValue)
}
