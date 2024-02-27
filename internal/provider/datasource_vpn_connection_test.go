package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccVPNConnectionDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVPNConnectionsConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func fetchVPNConnectionsConfig() string {
	return `data "numspot_load_balancers" "test" {}`
}
