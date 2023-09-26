package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNetResource(t *testing.T) {
	ipRange := "10.0.0.0/16"
	ipRangeUpdated := "10.0.0.1/16"
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetConfig_Create(ipRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net.test", "ip_range", ipRange),
					resource.TestCheckResourceAttrWith("numspot_net.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_net.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetConfig_Update(ipRangeUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net.test", "ip_range", ipRangeUpdated),
					resource.TestCheckResourceAttrWith("numspot_net.test", "id", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testNetConfig_Create(ipRange string) string {
	return fmt.Sprintf(`resource "numspot_net" "test" {
			ip_range=%s
  			}`, ipRange)
}

func testNetConfig_Update(ipRange string) string {
	return fmt.Sprintf(`resource "numspot_net" "test" {
			ip_range=%s
  			}`, ipRange)
}
