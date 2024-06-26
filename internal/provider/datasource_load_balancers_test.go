//go:build acc

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoadBalancerDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchLoadBalancersConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_load_balancers.test", "items.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_load_balancers.test", "items.0.name", "elb-test"),
				),
			},
		},
	})
}

func fetchLoadBalancersConfig() string {
	return `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "lb-test" {
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
}
data "numspot_load_balancers" "test" {
  load_balancer_names = [numspot_load_balancer.lb-test.name]
}`
}
