package test

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
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
func getDependencyChecksVpnConnection(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "client_gateway_id", "numspot_client_gateway.test"+dependenciesSuffix, "id"),   // If field is an id
		resource.TestCheckResourceAttrPair("numspot_vpn_connection.test", "virtual_gateway_id", "numspot_virtual_gateway.test"+dependenciesSuffix, "id"), // If field is an id
	}
}

func TestAccVpnConnectionResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	// The ASN used to create the client gateway
	// Each client gateway must use a distinct ASN
	// If your provide an ASN matching an existing gateway, the Create function will return the id of the existing one (thanks outscale)
	asn := 12345
	asnUpdated := 54321

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	routes := []string{"10.12.0.0/16"}
	routesUpdated := []string{"192.0.2.0/24", "192.168.255.0/24"}

	tunnelInsideIpRange := "169.254.254.22/30"
	tunnelInsideIpRangeUpdated := "169.254.254.20/30"

	presharedKey := "sample key !"
	presharedKeyUpdated := "another key !"

	tagKey := "Name"
	tagValue := "VPN-Connection-Test"
	tagValueUpdated := tagValue + "-Updated"

	// resource fields that cannot be updated in-place (requires replace)
	staticRouteOnly := "true"
	// staticRouteOnlyUpdated := "true"

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
		staticRoutesOnly:    staticRouteOnly,
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
				Config: testVpnConnectionConfig(acctest.BASE_SUFFIX, asn, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksVpnConnection(acctest.BASE_SUFFIX),
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
				Config: testVpnConnectionConfig(acctest.BASE_SUFFIX, asn, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksVpnConnection(acctest.BASE_SUFFIX),
				)...),
			},

			// <== If resource has required dependencies ==>
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testVpnConnectionConfig(acctest.NEW_SUFFIX, asnUpdated, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksVpnConnection(acctest.NEW_SUFFIX),
				)...),
			},
		},
	})
}

func testVpnConnectionConfig(subresourceSuffix string, asn int, data StepDataVpnConnection) string {
	routes := "["
	for _, route := range data.routes {
		routes += fmt.Sprintf("{destination_ip_range = %[1]q}", route)
		routes += ","
	}
	routes = strings.TrimSuffix(routes, ",")

	routes += "]"

	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_client_gateway" "test%[1]s" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = %[8]d
}

resource "numspot_virtual_gateway" "test%[1]s" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test%[1]s.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test%[1]s.id
  routes             = %[2]s
  static_routes_only = %[3]q
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
}`, subresourceSuffix,
		routes,
		data.staticRoutesOnly,
		data.tagKey,
		data.tagValue,
		data.preSharedKey,
		data.tunnelInsideIpRange,
		asn,
	)
}
