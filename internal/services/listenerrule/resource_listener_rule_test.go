//go:build acc

package listenerrule_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccListenerRuleResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testListenerRuleConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_listener_rule.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_listener_rule.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_listener_rule.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testListenerRuleConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_listener_rule.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_listener_rule.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testListenerRuleConfig_Create() string {
	return `resource "numspot_listener_rule" "test" {}`
}

func testListenerRuleConfig_Update() string {
	return `resource "numspot_listener_rule" "test" {
}`
}
