//go:build acc

package directlinkinterface_test

/*

 DIRECT LINKS are not handled for now

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccDirectLinkInterfaceResource(t *testing.T) {

	pr := provider.TestAccProtoV6ProviderFactories
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
*/
