//go:build acc

package clientgateway_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataClientGateway struct {
	tagKey,
	tagValue,
	bgpAsn,
	publicIp string
}

// Generate checks to validate that resource 'numspot_client_gateway.test' has input data values
func getFieldMatchChecksClientGateway(data StepDataClientGateway) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", data.publicIp),
		resource.TestCheckResourceAttr("numspot_client_gateway.test", "bgp_asn", data.bgpAsn),
		resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
	}
}

// Generate checks to validate that resource 'numspot_client_gateway.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksClientGateway(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{}
}

func TestAccClientGatewayResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "Name"
	tagValue := "Terraform-Test-Client-Gateway"
	tagValueUpdated := tagValue + "-GatewayUpdated"

	// resource fields that cannot be updated in-place (requires replace)
	bgpAsn := "65000"
	bgpAsnUpdated := "65001"
	publicIp := "192.0.2.0"
	publicIpUpdated := "192.0.3.0"
	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataClientGateway{
		tagKey:   tagKey,
		tagValue: tagValue,
		bgpAsn:   bgpAsn,
		publicIp: publicIp,
	}
	createChecks := append(
		getFieldMatchChecksClientGateway(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataClientGateway{
		tagKey:   tagKey,
		tagValue: tagValueUpdated,
		bgpAsn:   bgpAsn,
		publicIp: publicIp,
	}
	updateChecks := append(
		getFieldMatchChecksClientGateway(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataClientGateway{
		tagKey:   tagKey,
		tagValue: tagValue,
		bgpAsn:   bgpAsnUpdated,
		publicIp: publicIpUpdated,
	}
	replaceChecks := append(
		getFieldMatchChecksClientGateway(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
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
				Config: testClientGatewayConfig(provider.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksClientGateway(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_client_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testClientGatewayConfig(provider.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksClientGateway(provider.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testClientGatewayConfig(provider.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksClientGateway(provider.BASE_SUFFIX),
				)...),
			},
		},
	})
}

func testClientGatewayConfig(subresourceSuffix string, data StepDataClientGateway) string {
	return fmt.Sprintf(`
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = %[1]q
  bgp_asn         = %[2]s

  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, data.publicIp, data.bgpAsn, data.tagKey, data.tagValue)
}
