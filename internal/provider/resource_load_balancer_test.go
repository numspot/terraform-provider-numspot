package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccLoadBalancerResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testLoadBalancerConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_load_balancer.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testLoadBalancerConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testLoadBalancerConfig_Create() string {
	return fmt.Sprintf(`resource "numspot_load_balancer" "test" {
  			}`)
}

func testLoadBalancerConfig_Update() string {
	return `resource "numspot_load_balancer" "test" {
    			}`
}
