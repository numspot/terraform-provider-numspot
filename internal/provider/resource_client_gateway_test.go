//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccClientGatewayResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	connectionType := "ipsec.1"
	publicIp := "192.0.2.0"
	publicIpUpdate := "192.0.3.0"
	bgpAsn := 65000
	previousId := ""

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testClientGatewayConfig(connectionType, publicIp, bgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "connection_type", connectionType),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", publicIp),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "bgp_asn", func(v string) error {
						require.NotEmpty(t, v)
						require.Equal(t, fmt.Sprint(bgpAsn), v)
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						previousId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_client_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testClientGatewayConfig(connectionType, publicIpUpdate, bgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "connection_type", connectionType),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", publicIpUpdate),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "bgp_asn", func(v string) error {
						require.NotEmpty(t, v)
						require.Equal(t, fmt.Sprint(bgpAsn), v)
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						require.NotEqual(t, previousId, v)
						return nil
					}),
				),
			},
		},
	})
}

func testClientGatewayConfig(connectionType, publicIp string, bgpAsn int) string {
	return fmt.Sprintf(`resource "numspot_client_gateway" "test" {
  connection_type = %[1]q
  public_ip       = %[2]q
  bgp_asn         = %d
}`, connectionType, publicIp, bgpAsn)
}

func TestAccClientGatewayResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	connectionType := "ipsec.1"
	publicIp := "192.0.2.0"
	bgpAsn := 65000
	previousId := ""

	tagKey := "Name"
	tagValue := "Terraform-Test-Client-Gateway"
	tagValueUpdated := tagValue + "-GatewayUpdated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testClientGatewayConfig_Tags(connectionType, publicIp, bgpAsn, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "connection_type", connectionType),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", publicIp),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "bgp_asn", func(v string) error {
						require.NotEmpty(t, v)
						require.Equal(t, fmt.Sprint(bgpAsn), v)
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						previousId = v
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_client_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testClientGatewayConfig_Tags(connectionType, publicIp, bgpAsn, tagKey, tagValueUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "connection_type", connectionType),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", publicIp),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "bgp_asn", func(v string) error {
						require.NotEmpty(t, v)
						require.Equal(t, fmt.Sprint(bgpAsn), v)
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						require.Equal(t, previousId, v)
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.0.value", tagValueUpdated),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testClientGatewayConfig_Tags(connectionType, publicIp string, bgpAsn int, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_client_gateway" "test" {
  connection_type = %[1]q
  public_ip       = %[2]q
  bgp_asn         = %d

  tags = [
	{
	  key 		= %[4]q
	  value	 	= %[5]q
	}
  ]
}`, connectionType, publicIp, bgpAsn, tagKey, tagValue)
}
