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
			},
		},
	})
}

func fetchLoadBalancersConfig() string {
	return `data "numspot_load_balancers" "test" {}`
}
