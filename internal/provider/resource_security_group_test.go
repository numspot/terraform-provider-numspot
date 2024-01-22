package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccSecurityGroupResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	netId := "vpc-f1f48ebd"

	rand := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("security-group-name-%d", rand)
	descrition := fmt.Sprintf("security-group-description-%d", rand)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig(netId, name, descrition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", name),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", descrition),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSecurityGroupConfig(netId, name, descrition),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testSecurityGroupConfig(netId, name, description string) string {
	return fmt.Sprintf(`
resource "numspot_security_group" "test" {
	net_id = %[1]q
	name = %[2]q
	description = %[3]q
	inbound_rules = [
		{
			from_port_range = 80
			to_port_range = 80
			ip_ranges = ["0.0.0.0/0"]
			ip_protocol = "tcp"
		}
	]
}`, netId, name, description)
}
