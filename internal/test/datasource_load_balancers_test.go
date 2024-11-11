package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccLoadBalancerDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "test" {
  name = "elb-test"
  listeners = [
    {
      backend_port           = "80"
      load_balancer_port     = "80"
      load_balancer_protocol = "TCP"

    }
  ]
  subnets = [numspot_subnet.subnet.id]
  type    = "internal"
}
data "numspot_load_balancers" "testdata" {
  load_balancer_names = [numspot_load_balancer.test.name]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_load_balancers.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_load_balancers.testdata", "items.*", map[string]string{
						"subnets.0":                          acctest.PAIR_PREFIX + "numspot_subnet.subnet.id",
						"name":                               "elb-test",
						"listeners.0.backend_port":           "80",
						"listeners.0.load_balancer_port":     "80",
						"listeners.0.load_balancer_protocol": "TCP",
					}),
				),
			},
		},
	})
}
