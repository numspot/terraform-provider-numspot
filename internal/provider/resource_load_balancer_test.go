//go:build acc

package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoadBalancerResource(t *testing.T) {
	t.Parallel()
	randName := rand.Intn(9999-1000) + 1000
	lbName := fmt.Sprintf("elb-%d", randName)

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: loadBalancerResource_Config(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
				),
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

func loadBalancerResource_Config(name string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
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
}`, name)
}

func TestAccLoadBalancerResource_WithVm(t *testing.T) {
	t.Parallel()
	randName := rand.Intn(9999-1000) + 1000
	lbName := fmt.Sprintf("elb-%d", randName)

	imageId := "ami-8ef5b47e"

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: loadBalancerResource_WithVm_Config(lbName, imageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", lbName),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "backend_vm_ids.#", "1"),
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
