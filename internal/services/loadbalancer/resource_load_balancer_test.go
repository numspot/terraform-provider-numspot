//go:build acc

package loadbalancer_test

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This test sequence is quite long (~8 minutes)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataLoadBalancer struct {
	name,
	tagKey,
	tagValue string
	ports []string
}

// Generate checks to validate that resource 'numspot_load_balancer.test' has input data values
func getFieldMatchChecksLoadBalancer(data StepDataLoadBalancer) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", data.name),
		resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "tags.*", map[string]string{
			"key":   data.tagKey,
			"value": data.tagValue,
		}),
		resource.TestCheckResourceAttr("numspot_load_balancer.test", "listeners.#", strconv.Itoa(len(data.ports))),
	}

	for _, port := range data.ports {
		checks = append(checks, resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
			"backend_port":       port,
			"load_balancer_port": port,
		}))
	}

	return checks
}

// Generate checks to validate that resource 'numspot_load_balancer.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksLoadBalancer(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "subnets.*", "numspot_subnet.test"+dependenciesSuffix, "id"),
		resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "security_groups.*", "numspot_security_group.test"+dependenciesSuffix, "id"),
		resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "backend_vm_ids.*", "numspot_vm.test"+dependenciesSuffix, "id"),
	}
}

func TestAccLoadBalancerResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	subnetIpRange := "10.101.1.0/24"
	subnetIpRangeUpdated := "10.101.2.0/24"
	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	tagKey := "Name"
	tagValue := "ThisIsATerraformTest"
	tagValueUpdated := "ThisIsATerraformTestUpdated"

	ports := []string{"80"}
	portsUpdated := []string{"443", "8080"}

	// resource fields that cannot be updated in-place (requires replace)
	name := "elb-terraform-test"
	nameUpdated := "elb-terraform-test-updated"

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataLoadBalancer{
		tagKey:   tagKey,
		tagValue: tagValue,
		name:     name,
		ports:    ports,
	}
	createChecks := append(
		getFieldMatchChecksLoadBalancer(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataLoadBalancer{
		tagKey:   tagKey,
		tagValue: tagValueUpdated,
		name:     name,
		ports:    portsUpdated,
	}
	updateChecks := append(
		getFieldMatchChecksLoadBalancer(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)
	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataLoadBalancer{
		tagKey:   tagKey,
		tagValue: tagValue,
		name:     nameUpdated,
		ports:    ports,
	}
	replaceChecks := append(
		getFieldMatchChecksLoadBalancer(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
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
				Config: testLoadBalancerConfig(provider.BASE_SUFFIX, subnetIpRange, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksLoadBalancer(provider.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_load_balancer.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testLoadBalancerConfig(provider.BASE_SUFFIX, subnetIpRange, updatePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks,
					getDependencyChecksLoadBalancer(provider.BASE_SUFFIX),
				)...),
			},
			// Update testing With Replace (if needed)
			{
				Config: testLoadBalancerConfig(provider.BASE_SUFFIX, subnetIpRange, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksLoadBalancer(provider.BASE_SUFFIX),
				)...),
			},
			// Update from Internal Load Balancer to Public Load Balancer
			{
				Config: testLoadBalancerConfig_Public(provider.BASE_SUFFIX, subnetIpRange, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksLoadBalancer(provider.BASE_SUFFIX),
				)...),
			},
			// <== If resource has required dependencies ==>
			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: testLoadBalancerConfig(provider.BASE_SUFFIX, subnetIpRange, basePlanValues),
			},
			// --> DELETED TEST <-- : due to Numspot APIs architecture, this use case will not work in most cases. Nothing can be done on provider side to fix this
			// Update testing With Replace of dependency resource and without Replacing the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly

			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: testLoadBalancerConfig(provider.NEW_SUFFIX, subnetIpRangeUpdated, replacePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					replaceChecks,
					getDependencyChecksLoadBalancer(provider.NEW_SUFFIX),
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

func testLoadBalancerConfig(subresourceSuffix string, subnetIpRange string, data StepDataLoadBalancer) string {
	listeners := "["

	for _, port := range data.ports {
		listeners += fmt.Sprintf(`
    {
      backend_port           = %[1]s
      load_balancer_port     = %[1]s
      load_balancer_protocol = "TCP"
    }`, port)
		listeners += ","
	}
	listeners += "]"

	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test%[1]s" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = %[6]q
}

resource "numspot_security_group" "test%[1]s" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_vm" "test%[1]s" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-2c4r"

  subnet_id = numspot_subnet.test%[1]s.id
}

resource "numspot_load_balancer" "test" {
  name      = %[2]q
  listeners = %[5]s

  subnets         = [numspot_subnet.test%[1]s.id]
  security_groups = [numspot_security_group.test%[1]s.id]
  backend_vm_ids  = [numspot_vm.test%[1]s.id]

  type = "internal"

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]
}`, subresourceSuffix, data.name, data.tagKey, data.tagValue, listeners, subnetIpRange)
}

func testLoadBalancerConfig_Public(subresourceSuffix string, subnetIpRange string, data StepDataLoadBalancer) string {
	listeners := "["

	for _, port := range data.ports {
		listeners += fmt.Sprintf(`
    {
      backend_port           = %[1]s
      load_balancer_port     = %[1]s
      load_balancer_protocol = "TCP"
    }`, port)
		listeners += ","
	}
	listeners += "]"

	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "ig" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_subnet" "test%[1]s" {
  vpc_id                  = numspot_vpc.vpc.id
  ip_range                = %[6]q
  map_public_ip_on_launch = true
}

resource "numspot_security_group" "test%[1]s" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.vpc.id
  subnet_id = numspot_subnet.test%[1]s.id

  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.ig.id
    }
  ]
}

resource "numspot_vm" "test%[1]s" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-2c4r"

  subnet_id = numspot_subnet.test%[1]s.id
}

resource "numspot_load_balancer" "test" {
  name      = %[2]q
  listeners = %[5]s

  subnets         = [numspot_subnet.test%[1]s.id]
  security_groups = [numspot_security_group.test%[1]s.id]
  backend_vm_ids  = [numspot_vm.test%[1]s.id]

  type = "internet-facing"

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = %[3]q
      value = %[4]q
    }
  ]

  depends_on = [numspot_internet_gateway.ig]
}`, subresourceSuffix, data.name, data.tagKey, data.tagValue, listeners, subnetIpRange)
}

// <== If resource has optional dependencies ==>
func testLoadBalancerConfig_DeletedDependencies(data StepDataLoadBalancer) string {
	return fmt.Sprintf(`
resource "numspot_load_balancer" "test" {
  name = "%s"
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    }
  ]

  type = "internal"

  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, data.name, data.tagKey, data.tagValue)
}
