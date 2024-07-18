//go:build acc

package securitygroup_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccSecurityGroupsDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchSecurityGroupConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_security_groups.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_security_groups.testdata", "items.*", map[string]string{
						"id": provider.PAIR_PREFIX + "numspot_security_group.test.id",
					}),
				),
			},
		},
	})
}

func fetchSecurityGroupConfig() string {
	return `
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 443
      to_port_range   = 443
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

data "numspot_security_groups" "testdata" {
  ids        = [numspot_security_group.test.id]
  depends_on = [numspot_security_group.test]
}


`
}
