//go:build acc

package provider

import (
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataSecurityGroup struct {
	name, description, tagKey, tagValue   string
	inboundRulesPorts, outboundRulesPorts []string
}

// Generate checks to validate that resource 'numspot_security_group.test' has input data values
func getFieldMatchChecksSecurityGroup(data StepDataSecurityGroup) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_security_group.test", "name", data.name),
		resource.TestCheckResourceAttr("numspot_security_group.test", "description", data.description),
		resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
		resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.#", strconv.Itoa(len(data.inboundRulesPorts))),
		resource.TestCheckResourceAttr("numspot_security_group.test", "outbound_rules.#", strconv.Itoa(len(data.outboundRulesPorts))),
	}

	for _, inboundRulePort := range data.inboundRulesPorts {
		checks = append(checks, resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
			"from_port_range": inboundRulePort,
			"to_port_range":   inboundRulePort,
		}))
	}

	for _, outboundRulePort := range data.outboundRulesPorts {
		checks = append(checks, resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
			"from_port_range": outboundRulePort,
			"to_port_range":   outboundRulePort,
		}))
	}

	return checks
}

// Generate checks to validate that resource 'numspot_security_group.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksSecurityGroup(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test"+dependenciesPrefix, "id"),
	}
}

func TestAccSecurityGroupResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place

	inboundRulesPorts := []string{"453", "80", "22"}
	inboundRulesPortsUpdated_1 := []string{"453", "20"}
	inboundRulesPortsUpdated_2 := []string{}

	outboundRulesPorts := []string{"455", "90"}
	outboundRulesPortsUpdated_1 := []string{}
	outboundRulesPortsUpdated_2 := []string{"455", "90", "80", "70"}

	tagKey := "name"
	tagValue := "Terraform-Test-SecurityGroup"
	tagValueUpdated := "Terraform-Test-SecurityGroup-Updated"

	// resource fields that cannot be updated in-place (requires replace)
	randName := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("security-group-name-%d", randName)
	description := fmt.Sprintf("security-group-description-%d", randName)

	nameUpdated := name + "_updated"
	descriptionUpdated := description + "_updated"
	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataSecurityGroup{
		name:               name,
		description:        description,
		tagKey:             tagKey,
		tagValue:           tagValue,
		inboundRulesPorts:  inboundRulesPorts,
		outboundRulesPorts: outboundRulesPorts,
	}
	createChecks := append(
		getFieldMatchChecksSecurityGroup(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues_1 := StepDataSecurityGroup{
		name:               name,
		description:        description,
		tagKey:             tagKey,
		tagValue:           tagValueUpdated,
		inboundRulesPorts:  inboundRulesPortsUpdated_1,
		outboundRulesPorts: outboundRulesPortsUpdated_1,
	}
	updateChecks_1 := append(
		getFieldMatchChecksSecurityGroup(updatePlanValues_1),

		resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	updatePlanValues_2 := StepDataSecurityGroup{
		name:               name,
		description:        description,
		tagKey:             tagKey,
		tagValue:           tagValueUpdated,
		inboundRulesPorts:  inboundRulesPortsUpdated_2,
		outboundRulesPorts: outboundRulesPortsUpdated_2,
	}
	updateChecks_2 := append(
		getFieldMatchChecksSecurityGroup(updatePlanValues_2),

		resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataSecurityGroup{
		name:               nameUpdated,
		description:        descriptionUpdated,
		tagKey:             tagKey,
		tagValue:           tagValue,
		inboundRulesPorts:  inboundRulesPorts,
		outboundRulesPorts: outboundRulesPorts,
	}
	replaceChecks := append(
		getFieldMatchChecksSecurityGroup(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
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
				Config: testSecurityGroupConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksSecurityGroup(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testSecurityGroupConfig(utils_acctest.BASE_SUFFIX, updatePlanValues_1),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_1,
					getDependencyChecksSecurityGroup(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing Without Replace (if needed)
			{
				Config: testSecurityGroupConfig(utils_acctest.BASE_SUFFIX, updatePlanValues_2),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_2,
					getDependencyChecksSecurityGroup(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testSecurityGroupConfig(utils_acctest.BASE_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksSecurityGroup(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testSecurityGroupConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testSecurityGroupConfig(utils_acctest.NEW_SUFFIX, updatePlanValues_1),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_1,
					getDependencyChecksSecurityGroup(utils_acctest.NEW_SUFFIX),
				)...),
			},
			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testSecurityGroupConfig(utils_acctest.NEW_SUFFIX, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksSecurityGroup(utils_acctest.NEW_SUFFIX),
				)...),
			},

			// <== If resource has optional dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testSecurityGroupConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testSecurityGroupConfig_DeletedDependencies(updatePlanValues_1),
				Check:  resource.ComposeAggregateTestCheckFunc(updateChecks_1...),
			},
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testSecurityGroupConfig(utils_acctest.BASE_SUFFIX, basePlanValues),
			},
			// Update testing With Deletion of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly (empty dependency)
			// Note : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			{
				Config: testSecurityGroupConfig_DeletedDependencies(replacePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(replaceChecks...),
			},
		},
	})
}

func testSecurityGroupConfig(subresourceSuffix string, data StepDataSecurityGroup) string {
	inboundRules, outboundRules := getRules(data)

	return fmt.Sprintf(`
resource "numspot_vpc" "test%[1]s" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id         = numspot_vpc.test%[1]s.id
  name           = %[2]q
  description    = %[3]q
  inbound_rules  = %[4]s
  outbound_rules = %[5]s
}`, subresourceSuffix, data.name, data.description, inboundRules, outboundRules)
}

// <== If resource has optional dependencies ==>
func testSecurityGroupConfig_DeletedDependencies(data StepDataSecurityGroup) string {
	inboundRules, outboundRules := getRules(data)

	return fmt.Sprintf(`
resource "numspot_security_group" "test" {
  name           = %[1]q
  description    = %[2]q
  inbound_rules  = %[3]s
  outbound_rules = %[4]s
}`, data.name, data.description, inboundRules, outboundRules)
}

func getRules(data StepDataSecurityGroup) (string, string) {
	inboundRules := "["
	outboundRules := "["

	for _, port := range data.inboundRulesPorts {
		inboundRules += ruleFromPort(port)
		inboundRules += ","
	}

	for _, port := range data.outboundRulesPorts {
		outboundRules += ruleFromPort(port)
		outboundRules += ","
	}

	outboundRules = strings.TrimSuffix(outboundRules, ",")
	inboundRules = strings.TrimSuffix(inboundRules, ",")

	inboundRules += "]"
	outboundRules += "]"

	return inboundRules, outboundRules
}

func ruleFromPort(port string) string {
	return fmt.Sprintf(`{
		from_port_range = %[1]s
		to_port_range   = %[1]s
		ip_ranges       = ["0.0.0.0/0"]
		ip_protocol     = "tcp"
	  }`, port)
}
