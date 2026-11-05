//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccSecurityGroupRuleResource(t *testing.T) {
	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// 1️⃣ - Create Security group and outbound rule
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "sg" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "security-group-name"
  description = "security-group-description"
}

resource "numspot_security_group_rule" "security-group-rule_out" {
  security_group_id = numspot_security_group.sg.id
  flow              = "Outbound"
  from_port_range   = 1
  to_port_range     = 1
  ip_protocol       = "tcp"
  ip_range          = "10.0.0.0/16"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_out", "flow", "Outbound"),
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_out", "from_port_range", "1"),
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_out", "to_port_range", "1"),
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_out", "ip_protocol", "tcp"),
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_out", "ip_range", "10.0.0.0/16"),
				),
			},

			// 2️⃣ - Add Inbound rule
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "sg" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "security-group-name"
  description = "security-group-description"
}

resource "numspot_security_group_rule" "security-group-rule_in" {
  security_group_id = numspot_security_group.sg.id
  flow              = "Inbound"
  from_port_range   = 22
  to_port_range     = 22
  ip_protocol       = "tcp"
  ip_range          = "0.0.0.0/0"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_in", "flow", "Inbound"),
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_in", "from_port_range", "22"),
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_in", "to_port_range", "22"),
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_in", "ip_protocol", "tcp"),
					resource.TestCheckResourceAttr("numspot_security_group_rule.security-group-rule_in", "ip_range", "0.0.0.0/0"),
				),
			},
		},
	})
}
