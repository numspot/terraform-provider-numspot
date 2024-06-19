//go:build acc

package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoadBalancerResource(t *testing.T) {
	// t.Parallel()
	randName := rand.Intn(9999-1000) + 1000
	lbName := fmt.Sprintf("elb-%d", randName)

	listenerPort := 80
	updatedListenerPort := 443
	updated2ListenerPort := 8080

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: loadBalancerResource_Config(lbName, listenerPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.#", "1"),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.backend_port", fmt.Sprint(listenerPort)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.load_balancer_port", fmt.Sprint(listenerPort)),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_load_balancer.testlb",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"availability_zone_names"},
			},
			// Update listener port
			{
				Config: loadBalancerResource_Config(lbName, updatedListenerPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.#", "1"),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.backend_port", fmt.Sprint(updatedListenerPort)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.load_balancer_port", fmt.Sprint(updatedListenerPort)),
				),
			},
			// Add second listener
			{
				Config: loadBalancerResource_ConfigWithTwoListeners(lbName, updatedListenerPort, updated2ListenerPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.#", "2"),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.backend_port", fmt.Sprint(updatedListenerPort)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.load_balancer_port", fmt.Sprint(updatedListenerPort)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.1.backend_port", fmt.Sprint(updated2ListenerPort)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.1.load_balancer_port", fmt.Sprint(updated2ListenerPort)),
				),
			},
			// Add Health Check
			{
				Config: loadBalancerResource_ConfigWithHealthCheck(lbName, listenerPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.#", "1"),
				),
			},
			// First step state
			{
				Config: loadBalancerResource_Config(lbName, listenerPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.#", "1"),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.backend_port", fmt.Sprint(listenerPort)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.load_balancer_port", fmt.Sprint(listenerPort)),
				),
			},
			// Security groups
			{
				Config: loadBalancerResource_ConfigWithSecurityGroups(lbName, listenerPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_load_balancer.testlb", "security_groups.0", "numspot_security_group.sg", "id"),
				),
			},
			// Detach security groups
			{
				Config: loadBalancerResource_ConfigWithSecurityGroupsDetached(lbName, listenerPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.#", "1"),
				),
			},
			// First step state
			{
				Config: loadBalancerResource_Config(lbName, listenerPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.#", "1"),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.backend_port", fmt.Sprint(listenerPort)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "listeners.0.load_balancer_port", fmt.Sprint(listenerPort)),
				),
			},
		},
	})
}

func loadBalancerResource_Config(name string, listenerPort int) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
  name = %[1]q
  listeners = [
    {
      backend_port           = %[2]d
      load_balancer_port     = %[2]d
      load_balancer_protocol = "TCP"
    }
  ]

  subnets = [numspot_subnet.subnet.id]

  type = "internal"
}`, name, listenerPort)
}

func loadBalancerResource_ConfigWithTwoListeners(name string, listenerPortA, listenerPortB int) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
  name = %[1]q

  listeners = [
    {
      backend_port           = %[2]d
      load_balancer_port     = %[2]d
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = %[3]d
      load_balancer_port     = %[3]d
      load_balancer_protocol = "TCP"
    }
  ]

  subnets = [numspot_subnet.subnet.id]

  type = "internal"
}`, name, listenerPortA, listenerPortB)
}

func loadBalancerResource_ConfigWithHealthCheck(name string, listenerPort int) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
  name = %[1]q
  listeners = [
    {
      backend_port           = %[2]d
      load_balancer_port     = %[2]d
      load_balancer_protocol = "TCP"
    }
  ]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  subnets = [numspot_subnet.subnet.id]

  type = "internal"
}`, name, listenerPort)
}

func loadBalancerResource_ConfigWithSecurityGroups(name string, listenerPort int) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
  name = %[1]q
  listeners = [
    {
      backend_port           = %[2]d
      load_balancer_port     = %[2]d
      load_balancer_protocol = "TCP"
    }
  ]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  subnets = [numspot_subnet.subnet.id]

  security_groups = [numspot_security_group.sg.id]

  depends_on = [numspot_security_group.sg]

  type = "internal"
}

resource "numspot_security_group" "sg" {
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

  lifecycle {
    create_before_destroy = true
  }
}`, name, listenerPort)
}

func loadBalancerResource_ConfigWithSecurityGroupsDetached(name string, listenerPort int) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
  name = %[1]q
  listeners = [
    {
      backend_port           = %[2]d
      load_balancer_port     = %[2]d
      load_balancer_protocol = "TCP"
    }
  ]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  subnets = [numspot_subnet.subnet.id]

  depends_on = [numspot_security_group.sg]

  type = "internal"
}

resource "numspot_security_group" "sg" {
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

  lifecycle {
    create_before_destroy = true
  }
}`, name, listenerPort)
}

func TestAccLoadBalancerResource_WithVm(t *testing.T) {
	t.Parallel()
	randName := rand.Intn(9999-1000) + 1000
	lbName := fmt.Sprintf("elb-%d", randName)

	imageId := "ami-8ef5b47e"
	// vmType := "tinav6.c1r1p3"

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: loadBalancerResource_WithVm_Config(lbName, imageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
				),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_load_balancer.testlb",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"availability_zone_names"},
			},
		},
	})
}

func loadBalancerResource_WithVm_Config(name, imageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_vm" "example" {
  image_id = %[1]q
  type     = %[2]q

  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_load_balancer" "testlb" {
  name = %[3]q
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    }
  ]

  subnets        = [numspot_subnet.subnet.id]
  backend_vm_ids = [numspot_vm.example.id]

  type = "internal"
}`, imageId, vmType, name)
}

func TestAccLoadBalancerResource_PublicWithVm(t *testing.T) {
	t.Parallel()
	randName := rand.Intn(9999-1000) + 1000
	lbName := fmt.Sprintf("elb-%d", randName)

	imageId := "ami-8ef5b47e"
	// vmType := "tinav6.c1r1p3"

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: loadBalancerResource_PublicWithVm_Config(lbName, imageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
				),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_load_balancer.testlb",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"availability_zone_names", "public_ip"},
			},
		},
	})
}

func loadBalancerResource_PublicWithVm_Config(name, imageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"

  tags = [
    {
      key   = "env"
      value = "Terraform-Tests"
    }
  ]
}

resource "numspot_internet_gateway" "ig" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_subnet" "subnet" {
  vpc_id                  = numspot_vpc.vpc.id
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = true

  tags = [
    {
      key   = "env"
      value = "Terraform-Tests"
    }
  ]
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.vpc.id
  subnet_id = numspot_subnet.subnet.id

  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.ig.id
    }
  ]
}

resource "numspot_vm" "example" {
  image_id = %[1]q
  type     = %[2]q

  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_load_balancer" "testlb" {
  name = %[3]q
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    }
  ]

  subnets        = [numspot_subnet.subnet.id]
  backend_vm_ids = [numspot_vm.example.id]

  type = "internet-facing"
}`, imageId, vmType, name)
}

func TestAccLoadBalancerResource_Tags(t *testing.T) {
	t.Parallel()
	randName := rand.Intn(9999-1000) + 1000
	lbName := fmt.Sprintf("elb-%d", randName)

	tagKey := "Name"
	tagValue := "ThisIsATerraformTest"
	tagValueUpdate := "ThisIsATerraformTestUpdated"

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: loadBalancerResource_Config_Tags(lbName, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_load_balancer.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"availability_zone_names"},
			},
			{
				Config: loadBalancerResource_Config_Tags(lbName, tagKey, tagValueUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.0.value", tagValueUpdate),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
				),
			},
		},
	})
}

func loadBalancerResource_Config_Tags(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "test" {
  name = "%s"
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    }
  ]

  subnets = [numspot_subnet.subnet.id]

  type = "internal"

  tags = [
    {
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, name, tagKey, tagValue)
}
