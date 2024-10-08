package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccClientGatewayResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Step 1 - Create client gateway
			{
				Config: `
			resource "numspot_client_gateway" "terraform-client-gateway" {
			 connection_type = "ipsec.1"
			 public_ip       = "192.0.0.1"
			 bgp_asn         = 65000
			
			 tags = [
			   {
			     key   = "name"
			     value = "terraform-client-gateway"
			   }
			 ]
			}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "public_ip", "192.0.0.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "bgp_asn", "65000"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.terraform-client-gateway", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-client-gateway",
					}),
				),
			},
			// Step 2 - Import client gateway
			{
				ResourceName:            "numspot_client_gateway.terraform-client-gateway",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Step 3 - Replace client gateway public IP
			{
				Config: `
			resource "numspot_client_gateway" "terraform-client-gateway" {
			 connection_type = "ipsec.1"
			 public_ip       = "192.0.0.2"
			 bgp_asn         = 65000
			
			 tags = [
			   {
			     key   = "name"
			     value = "terraform-client-gateway"
			   }
			 ]
			}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "public_ip", "192.0.0.2"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "bgp_asn", "65000"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.terraform-client-gateway", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-client-gateway",
					}),
				),
			},
			// Step 4 - Replace client gateway BGP
			{
				Config: `
			resource "numspot_client_gateway" "terraform-client-gateway" {
			 connection_type = "ipsec.1"
			 public_ip       = "192.0.0.1"
			 bgp_asn         = 65001
			
			 tags = [
			   {
			     key   = "name"
			     value = "terraform-client-gateway"
			   }
			 ]
			}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "public_ip", "192.0.0.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "bgp_asn", "65001"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.terraform-client-gateway", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-client-gateway",
					}),
				),
			},
			// Step 5 - Delete client gateway
			{
				Config: ` `,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Step 6 - Create client gateway
			{
				Config: `
			resource "numspot_client_gateway" "terraform-client-gateway" {
			 connection_type = "ipsec.1"
			 public_ip       = "192.0.0.1"
			 bgp_asn         = "65000"
			
			 tags = [
			   {
			     key   = "name"
			     value = "terraform-client-gateway"
			   }
			 ]
			}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "public_ip", "192.0.0.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "bgp_asn", "65000"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.terraform-client-gateway", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-client-gateway",
					}),
				),
			},
			// Step 7 - Update client gateway tags
			{
				Config: `
resource "numspot_client_gateway" "terraform-client-gateway" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.0.1"
  bgp_asn         = "65000"

  tags = [
    {
      key   = "name"
      value = "terraform-client-gateway-update"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "public_ip", "192.0.0.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "bgp_asn", "65000"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.terraform-client-gateway", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-client-gateway-update",
					}),
				),
			},
			// Step 8 - Recreate client gateway
			{
				Config: `
resource "numspot_client_gateway" "terraform-client-gateway-recreate" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.0.1"
  bgp_asn         = "65000"

  tags = [
    {
      key   = "name"
      value = "terraform-client-gateway-recreate"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway-recreate", "connection_type", "ipsec.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway-recreate", "public_ip", "192.0.0.1"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway-recreate", "bgp_asn", "65000"),
					resource.TestCheckResourceAttr("numspot_client_gateway.terraform-client-gateway-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.terraform-client-gateway-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-client-gateway-recreate",
					}),
				),
			},
		},
	})
}
