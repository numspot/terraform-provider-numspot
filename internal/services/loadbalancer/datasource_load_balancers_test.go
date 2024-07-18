//go:build acc

package loadbalancer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccLoadBalancerDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	name := "elb-test"
	backend_port := "80"
	load_balancer_port := "80"
	load_balancer_protocol := "TCP"
	lb_type := "internal"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchLoadBalancersConfig(name, backend_port, load_balancer_port, load_balancer_protocol, lb_type),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_load_balancers.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_load_balancers.testdata", "items.*", map[string]string{
						"id":                                 provider.PAIR_PREFIX + "numspot_load_balancer.test.id",
						"subnets.0":                          provider.PAIR_PREFIX + "numspot_subnet.subnet.id",
						"name":                               name,
						"listeners.0.backend_port":           backend_port,
						"listeners.0.load_balancer_port":     load_balancer_port,
						"listeners.0.load_balancer_protocol": load_balancer_protocol,
					}),
				),
			},
		},
	})
}

func fetchLoadBalancersConfig(name, backend_port, load_balancer_port, load_balancer_protocol, lb_type string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "test" {
  name = %[1]q
  listeners = [
    {
      backend_port           = %[2]s
      load_balancer_port     = %[3]s
      load_balancer_protocol = %[4]q

    }
  ]
  subnets = [numspot_subnet.subnet.id]
  type    = %[5]q
}
data "numspot_load_balancers" "testdata" {
  load_balancer_names = [numspot_load_balancer.test.name]
}`, name, backend_port, load_balancer_port, load_balancer_protocol, lb_type)
}
