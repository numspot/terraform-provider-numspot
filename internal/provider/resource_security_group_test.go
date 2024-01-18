package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccSecurityGroupResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig("name-2", "description-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "field", func(v string) error {
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
				Config: testSecurityGroupConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testSecurityGroupConfig(name, description string) string {
	return fmt.Sprintf(`
resource "numspot_security_group" "test" {
	name = %[1]q
	description = %[2]q
	inbound_rules = [
		{
			from_port_range = 22
			to_port_range = 22
			ip_ranges = ["0.0.0.0/0"]
			ip_protocol = "-1"
		}
	]
}`, name, description)
}

func testSecurityGroupConfig_Update() string {
	return `resource "numspot_security_group" "test" {
    			}`
}
