//go:build acc

package natgateway_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataNatGateway struct {
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_nat_gateway.test' has input data values
func getFieldMatchChecksNatGateway(data StepDataNatGateway) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_nat_gateway.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_nat_gateway.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_nat_gateway.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksNatGateway(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "subnet_id", "numspot_subnet.test"+dependenciesSuffix, "id"),
		resource.TestCheckResourceAttrPair("numspot_nat_gateway.test", "public_ip_id", "numspot_public_ip.test"+dependenciesSuffix, "id"),
	}
}

func TestAccNatGatewayResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "name"
	tagValue := "Terraform-Test-Nat-Gateway"
	tagValueUpdated := tagValue + "-Updated"

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataNatGateway{
		tagKey:   tagKey,
		tagValue: tagValue,
	}
	createChecks := append(
		getFieldMatchChecksNatGateway(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_nat_gateway.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataNatGateway{
		tagKey:   tagKey,
		tagValue: tagValueUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksNatGateway(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_nat_gateway.test", "id", func(v string) error {
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
				Config: testNatGatewayConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksNatGateway(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nat_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testNatGatewayConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksNatGateway(provider.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			// --> DELETED TEST <-- : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
		},
	})
}

func testNatGatewayConfig(subresourceSuffix string, data StepDataNatGateway) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_subnet" "test%[1]s" {
  vpc_id                  = numspot_vpc.test.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "test%[1]s" {}

resource "numspot_route_table" "test%[1]s" {
  vpc_id    = numspot_vpc.test.id
  subnet_id = numspot_subnet.test%[1]s.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

resource "numspot_nat_gateway" "test" {
  subnet_id    = numspot_subnet.test%[1]s.id
  public_ip_id = numspot_public_ip.test%[1]s.id
  depends_on   = [numspot_route_table.test%[1]s]
  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}
`, subresourceSuffix, data.tagKey, data.tagValue)
}
