//go:build acc

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
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
      key   = %[2]q
      value = %[3]q
    }
  ]
}`, connectionType, tagKey, tagValue)
}

func TestAccVirtualGatewayResourceLink(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	var vg_id string

	vpcSuffix := "1"
	vpcSuffixUpdated := "2"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVirtualGatewayConfigLink(vpcSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						vg_id = v
						return nil
					}),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", fmt.Sprintf("numspot_vpc.vpc%v", vpcSuffix), "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_virtual_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testVirtualGatewayConfigLink(vpcSuffixUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_virtual_gateway.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						if vg_id != v {
							return errors.New("Id should be the same after Update without replace")
						}
						return nil
					}),
					resource.TestCheckResourceAttrPair("numspot_virtual_gateway.test", "vpc_id", fmt.Sprintf("numspot_vpc.vpc%v", vpcSuffixUpdated), "id"),
				),
			},
		},
	})
}

func testVirtualGatewayConfigLink(vpcSuffix string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc1" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_vpc" "vpc2" {
  ip_range = "10.121.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.vpc%[1]s.id
}
`, vpcSuffix)
}
