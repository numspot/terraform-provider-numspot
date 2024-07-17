//go:build acc

package provider

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataVpnConnection struct {
	routes []string
	staticRoutesOnly,
	preSharedKey,
	tunnelInsideIpRange,
	tagKey,
	tagValue string
}

// Generate checks to validate that resource 'numspot_vpn_connection.test' has input data values
func getFieldMatchChecksVpnConnection(data StepDataVpnConnection) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_vpn_connection.test", "static_routes_only", data.staticRoutesOnly),

		resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.phase2options.pre_shared_key", data.preSharedKey),
		resource.TestCheckResourceAttr("numspot_vpn_connection.test", "vpn_options.tunnel_inside_ip_range", data.tunnelInsideIpRange),
		resource.TestCheckResourceAttr("numspot_vpn_connection.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
		resource.TestCheckResourceAttr("numspot_vpn_connection.test", "routes.#", strconv.Itoa(len(data.routes))),
	}

	for _, route := range data.routes {
		checks = append(checks, resource.TestCheckTypeSetElemNestedAttrs("numspot_vpn_connection.test", "routes.*", map[string]string{
			"destination_ip_range": route,
		}))
	}

	return checks
}

// Generate checks to validate that resource 'numspot_vpn_connection.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksVpnConnection(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test"+dependenciesPrefix, "id"),   // If field is an id
		resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test"+dependenciesPrefix, "id"), // If field is an id
	}
}

func TestAccVpnConnectionResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	routes := []string{"10.12.0.0/16"}
	routesUpdated := []string{"192.0.2.0/24", "192.168.255.0/24"}

	tunnelInsideIpRange := "169.254.254.22/30"
	tunnelInsideIpRangeUpdated := "169.254.254.20/30"

	presharedKey := "sample key !"
	presharedKeyUpdated := "another key !"

	tagKey := "Name"
	tagValue := "Terraform-Test-Volume"
	tagValueUpdated := tagValue + "-Updated"

	// resource fields that cannot be updated in-place (requires replace)
	staticRouteOnly := "false"
	staticRouteOnlyUpdated := "true"

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataVpnConnection{
		staticRoutesOnly:    staticRouteOnly,
		routes:              routes,
		preSharedKey:        presharedKey,
		tunnelInsideIpRange: tunnelInsideIpRange,
		tagKey:              tagKey,
		tagValue:            tagValue,
	}
	createChecks := append(
		getFieldMatchChecksVpnConnection(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataVpnConnection{
		staticRoutesOnly:    staticRouteOnly,
		routes:              routesUpdated,
		preSharedKey:        presharedKeyUpdated,
		tunnelInsideIpRange: tunnelInsideIpRangeUpdated,
		tagKey:              tagKey,
		tagValue:            tagValueUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksVpnConnection(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataVpnConnection{
		staticRoutesOnly:    staticRouteOnlyUpdated,
		routes:              routes,
		preSharedKey:        presharedKey,
		tunnelInsideIpRange: tunnelInsideIpRange,
		tagKey:              tagKey,
		tagValue:            tagValue,
	}
	replaceChecks := append(
		getFieldMatchChecksVpnConnection(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_vpn_connection.test", "id", func(v string) error {
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
				Config: testVpnConnectionConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksVpnConnection(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpn_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testVpnConnectionConfig(utils_acctest.BASE_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVpnConnection(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testVpnConnectionConfig(utils_acctest.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksVpnConnection(utils_acctest.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testVpnConnectionConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testVpnConnectionConfig(utils_acctest.NEW_SUFFIX, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVpnConnection(utils_acctest.NEW_SUFFIX),
				)...),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testVpnConnectionConfig(utils_acctest.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksVpnConnection(utils_acctest.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func testVpnConnectionConfig(subresourceSuffix string, data StepDataVpnConnection) string {
	routes := "["
	for _, route := range data.routes {
		routes += fmt.Sprintf("{destination_ip_range = %[1]q}", route)
		routes += ","
	}
	routes = strings.TrimSuffix(routes, ",")

	routes += "]"

	return fmt.Sprintf(`
resource "numspot_client_gateway" "test%[1]s" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_virtual_gateway" "test%[1]s" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test%[1]s.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test%[1]s.id
  static_routes_only = %[3]q
  routes             = %[2]s
  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
  vpn_options = {
    phase2options = {
      pre_shared_key = %[6]q
    }
    tunnel_inside_ip_range = %[7]q
  }
}`, subresourceSuffix, routes, data.staticRoutesOnly, data.tagKey, data.tagValue, data.preSharedKey, data.tunnelInsideIpRange)
}
