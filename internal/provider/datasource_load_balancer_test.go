package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccLoadBalancerDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchLoadBalancersConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(),
				//resource.TestCheckResourceAttrSet(),
				//resource.TestCheckResourceAttr("hashicups_order.test", "items.#", "1")),
			},
		},
	})
}

func fetchLoadBalancersConfig() string {
	return `
data "numspot_load_balancers" "test" {
	load_balancer_names = [numspot_load_balancer.testlb.name]
}
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
	vpc_id 		= numspot_vpc.vpc.id
	ip_range 	= "10.101.1.0/24"
}

resource "numspot_security_group" "sg" {
	net_id 		= numspot_vpc.vpc.id
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
}`
}
