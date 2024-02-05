package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccDirectLinkInterfaceResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testDirectLinkInterfaceConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_direct_link_interface.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_direct_link_interface.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_direct_link_interface.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testDirectLinkInterfaceConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_direct_link_interface.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_direct_link_interface.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testDirectLinkInterfaceConfig_Create() string {
	return `resource "numspot_direct_link_interface" "test" {}`
}

func testDirectLinkInterfaceConfig_Update() string {
	return `resource "numspot_direct_link_interface" "test" {
    			}`
}
