package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNetAccessPointResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetAccessPointConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_net_access_point.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_net_access_point.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetAccessPointConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_net_access_point.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testNetAccessPointConfig_Create() string {
	return `resource "numspot_net_access_point" "test" {}`
}

func testNetAccessPointConfig_Update() string {
	return `resource "numspot_net_access_point" "test" {
    			}`
}
