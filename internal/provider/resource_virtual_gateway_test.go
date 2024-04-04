//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVirtualGatewayResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	connectionType := "ipsec.1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVirtualGatewayConfig(connectionType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", connectionType),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_virtual_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testVirtualGatewayConfig(connectionType string) string {
	return fmt.Sprintf(`resource "numspot_virtual_gateway" "test" {
  connection_type = %[1]q
}`, connectionType)
}
