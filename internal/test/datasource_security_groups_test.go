package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccSecurityGroupsDatasource(t *testing.T) {
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
  security_group_ids = [numspot_security_group.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_security_groups.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_security_groups.testdata", "items.*", map[string]string{
						"id": acctest.PAIR_PREFIX + "numspot_security_group.test.id",
					}),
				),
			},
		},
	})
}
