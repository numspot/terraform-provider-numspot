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
resource "numspot_vpc" "terraform-dep-vpc-load-balancer" {
  ip_range = "10.101.0.0/16"
  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest"
  }]
}

resource "numspot_subnet" "terraform-dep-subnet-load-balancer" {
  vpc_id   = numspot_vpc.terraform-dep-vpc-load-balancer.id
  ip_range = "10.101.1.0/24"
  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest"
  }]
}

resource "numspot_load_balancer" "terraform-load-balancer-acctest" {
  name = "terraform-load-balancer-acctest"
  listeners = [{
    backend_port           = "80"
    load_balancer_port     = "80"
    load_balancer_protocol = "TCP"

  }]
  subnets = [numspot_subnet.terraform-dep-subnet-load-balancer.id]
  type    = "internal"
  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest"
  }]
}

data "numspot_load_balancers" "datasource-load-balancer-acctest" {
  load_balancer_names = [numspot_load_balancer.terraform-load-balancer-acctest.name]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_load_balancers.datasource-load-balancer-acctest", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_load_balancers.datasource-load-balancer-acctest", "items.*", map[string]string{
						"name":                               "terraform-load-balancer-acctest",
						"listeners.0.backend_port":           "80",
						"listeners.0.load_balancer_port":     "80",
						"listeners.0.load_balancer_protocol": "TCP",
					}),
				),
			},
		},
	})
}
