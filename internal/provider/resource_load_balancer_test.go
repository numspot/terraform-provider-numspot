package provider

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoadBalancerResource(t *testing.T) {
	lbName := "elb-test"
	hc := api.HealthCheck{
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
				ImportStateVerifyIgnore: []string{},
			},
			//Update testing
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
					resource.TestCheckResourceAttrWith("numspot_load_balancer.testlb", "backend_vm_ids", func(v string) error {
						if assert.Empty(t, v) {
							return fmt.Errorf("backend_vm_ids attr empty after link to backend machines call")
						}
						return nil
					}),
				),
			},
		},
	})
}

func createLbConfig(name string) string {
	return fmt.Sprintf(`
resource "numspot_net" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
	net_id 		= numspot_net.net.id
	ip_range 	= "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
	name = "%s"
	listeners = [
		{
			backend_port = 80
			load_balancer_port = 80
			load_balancer_protocol = "TCP"
					
		}
	]
	subnets = [numspot_subnet.subnet.id]
	type = "internal"
}`, name)
}

func updateLbConfig(hc api.HealthCheck) string {
	return fmt.Sprintf(`
resource "numspot_net" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
	net_id 		= numspot_net.net.id
	ip_range 	= "10.101.1.0/24"
}

resource "numspot_load_balancer" "testlb" {
	name = "elb-test"
	listeners = [
		{
			backend_port = 80
			load_balancer_port = 80
			load_balancer_protocol = "TCP"
					
		}
	]
	subnets = [numspot_subnet.subnet.id]
	type = "internal"
	health_check = {
		check_interval = %d
		healthy_threshold = %d
		path = "%s"
		port = %d
		protocol = "%s"
		timeout = %d
		unhealthy_threshold = %d
	}
}`, hc.CheckInterval, hc.HealthyThreshold, *hc.Path, hc.Port, hc.Protocol, hc.Timeout, hc.UnhealthyThreshold)
}

func linkBackendMachinesToLbConfig() string {
	return fmt.Sprintf(`
resource "numspot_net" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
	net_id 		= numspot_net.net.id
	ip_range 	= "10.101.1.0/24"
}

resource "numspot_security_group" "sg" {
	net_id 		= numspot_net.net.id
	name 		= "terraform-vm-tests-sg-name"
	description = "terraform-vm-tests-sg-description"

	inbound_rules = [
		{
			from_port_range = 80
			to_port_range = 80
			ip_ranges = ["0.0.0.0/0"]
			ip_protocol = "tcp"
		}
	]
}

resource "numspot_internet_service" "is" {
  net_id = numspot_net.net.id
}

resource "numspot_route_table" "rt" {
  net_id    = numspot_net.net.id
  subnet_id = numspot_subnet.subnet.id

  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_service.is.id
    }
  ]
}

resource "numspot_vm" "test" {
	image_id 			= "ami-00b0c39a"
	vm_type 			= "t2.small"
	subnet_id			= numspot_subnet.subnet.id
	security_group_ids 	= [ numspot_security_group.sg.id ]
	depends_on 			= [ numspot_security_group.sg ]
}

resource "numspot_load_balancer" "testlb" {
	name = "elb-test"
	listeners = [
		{
			backend_port = 80
			load_balancer_port = 80
			load_balancer_protocol = "TCP"
					
		}
	]
	subnets = [numspot_subnet.subnet.id]
	type = "internal"
	health_check = {
		check_interval = 30
		healthy_threshold = 10
		path = "/index.html"
		port = 8080
		protocol = "HTTPS"
		timeout = 5
		unhealthy_threshold = 5
	}
	backend_vm_ids = [numspot_vm.test.id]
}`)
}
