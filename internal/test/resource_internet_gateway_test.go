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
type StepDataInternetGateway struct {
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_internet_gateway.test' has input data values
func getFieldMatchChecksInternetGateway(data StepDataInternetGateway) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_internet_gateway.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_internet_gateway.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_internet_gateway.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksInternetGateway(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_internet_gateway.test", "vpc_id", "numspot_vpc.test"+dependenciesSuffix, "id"),
	}
}

func TestAccInternetGatewayResource(t *testing.T) {
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
	tagValue := "Terraform-Test-Volume"
	tagValueUpdated := "Terraform-Test-Volume-Updated"

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataInternetGateway{
		tagKey:   tagKey,
		tagValue: tagValue,
	}
	createChecks := append(
		getFieldMatchChecksInternetGateway(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_internet_gateway.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataInternetGateway{
		tagKey:   tagKey,
		tagValue: tagValueUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksInternetGateway(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_internet_gateway.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.

	replaceChecks := append(
		getFieldMatchChecksInternetGateway(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_internet_gateway.test", "id", func(v string) error {
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
				Config: testInternetGatewayConfig(acctest.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksInternetGateway(acctest.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_internet_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testInternetGatewayConfig(acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksInternetGateway(acctest.BASE_SUFFIX),
				)...),
			},
			// <== If resource has required dependencies ==>
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			{
				Config: testInternetGatewayConfig(acctest.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksInternetGateway(acctest.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func testInternetGatewayConfig(subresourceSuffix string, data StepDataInternetGateway) string {
	return fmt.Sprintf(`

// <== If resource has dependencies ==> 
resource "numspot_vpc" "test%[4]s" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-internetgateway-acctest"
    }
  ]
}

resource "numspot_vpc" "test%[5]s" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-internetgateway-acctest"
    }
  ]
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test%[1]s.id
  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, subresourceSuffix, data.tagKey, data.tagValue, acctest.BASE_SUFFIX, acctest.NEW_SUFFIX)
}
