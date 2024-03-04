package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccDirectLinkResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testDirectLinkConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_direct_link.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_direct_link.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_direct_link.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			//{
			//	Config: testDirectLinkConfig_Update(),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("numspot_direct_link.test", "field", "value"),
			//		resource.TestCheckResourceAttrWith("numspot_direct_link.test", "field", func(v string) error {
			//			return nil
			//		}),
			//	),
			//},
		},
	})
}

func testDirectLinkConfig() string {
	return `
resource "numspot_direct_link" "test" {
	location = "PAR1"
	bandwidth = "1Gbps"
	name = "Connection to NumSpot"
}
`
}
