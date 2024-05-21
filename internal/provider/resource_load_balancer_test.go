//go:build acc

package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func TestAccLoadBalancerResource(t *testing.T) {
	t.Parallel()
	lbName := "elb-test"
	hc := iaas.HealthCheck{
		CheckInterval:      30,
		HealthyThreshold:   10,
		Path:               utils.PointerOf("/index.html"),
		Port:               8080,
		Protocol:           "HTTPS",
		Timeout:            5,
		UnhealthyThreshold: 5,
	}

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: createLbConfig(lbName),
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
			// Update testing
			{
				Config: updateLbConfig(hc),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.check_interval", strconv.Itoa(hc.CheckInterval)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.healthy_threshold", strconv.Itoa(hc.HealthyThreshold)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.path", *hc.Path),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.port", strconv.Itoa(hc.Port)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.protocol", hc.Protocol),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.timeout", strconv.Itoa(hc.Timeout)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.unhealthy_threshold", strconv.Itoa(hc.UnhealthyThreshold)),
					//resource.TestCheckResourceAttrWith("numspot_load_balancer.testlb", "field", func(v string) error {
					//	return nil
					//}),
				),
			},
			{
				Config: linkBackendMachinesToLbConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "backend_vm_ids.#", "1"),
				),
			},
		},
	})
}

func createLbConfig(name string) string {
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
  type    = "internal"
}`, name)
}

func updateLbConfig(hc iaas.HealthCheck) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
  name = "elb-test"
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"

    }
  ]
  subnets = [numspot_subnet.subnet.id]
  type    = "internal"
  health_check = {
    check_interval      = %d
    healthy_threshold   = %d
    path                = "%s"
    port                = %d
    protocol            = "%s"
    timeout             = %d
    unhealthy_threshold = %d
  }
}`, hc.CheckInterval, hc.HealthyThreshold, *hc.Path, hc.Port, hc.Protocol, hc.Timeout, hc.UnhealthyThreshold)
}

func linkBackendMachinesToLbConfig() string {
	return `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "sg" {
  net_id      = numspot_vpc.vpc.id
  name        = "terraform-vm-tests-sg-name"
  description = "terraform-vm-tests-sg-description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_internet_gateway" "igw" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_route_table" "rt" {
  net_id    = numspot_vpc.vpc.id
  subnet_id = numspot_subnet.subnet.id

  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.igw.id
    }
  ]
}

resource "numspot_vm" "test" {
  image_id           = "ami-00b0c39a"
  vm_type            = "ns-cus6-2c4r"
  subnet_id          = numspot_subnet.subnet.id
  security_group_ids = [numspot_security_group.sg.id]
  depends_on         = [numspot_security_group.sg]
}

resource "numspot_load_balancer" "testlb" {
  name = "elb-test"
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"

    }
  ]
  subnets = [numspot_subnet.subnet.id]
  type    = "internal"
  health_check = {
    check_interval      = 30
    healthy_threshold   = 10
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }
  backend_vm_ids = [numspot_vm.test.id]
}`
}
