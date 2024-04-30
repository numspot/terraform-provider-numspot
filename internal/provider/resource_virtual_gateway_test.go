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

func TestAccVirtualGatewayResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	connectionType := "ipsec.1"
	tagKey := "name"
	tagValue := "Terraform-Test-Volume"
	tagValueUpdate := tagValue + "-Update"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVirtualGatewayConfig_Tags(connectionType, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "connection_type", connectionType),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_virtual_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testVirtualGatewayConfig_Tags(connectionType, tagKey, tagValueUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.0.value", tagValueUpdate),
					resource.TestCheckResourceAttr("numspot_virtual_gateway.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testVirtualGatewayConfig_Tags(connectionType, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_virtual_gateway" "test" {
  connection_type = %[1]q

  tags = [
	{
	  key 		= %[2]q
	  value	 	= %[3]q
	}
  ]
}`, connectionType, tagKey, tagValue)
}
