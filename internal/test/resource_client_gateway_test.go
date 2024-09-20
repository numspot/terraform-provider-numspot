package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
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

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create testing
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = "65000"

  tags = [
    {
      key   = "Name"
      value = "Terraform-Test-Client-Gateway"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", "192.0.2.0"),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "bgp_asn", "65000"),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.test", "tags.*", map[string]string{
						"key":   "Name",
						"value": "Terraform-Test-Client-Gateway",
					}),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_client_gateway.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = "65000"

  tags = [
    {
      key   = "Name"
      value = "Terraform-Test-Client-GatewayUpdated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", "192.0.2.0"),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "bgp_asn", "65000"),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.test", "tags.*", map[string]string{
						"key":   "Name",
						"value": "Terraform-Test-Client-GatewayUpdated",
					}),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
			},
			// 4 - Update testing With Replace (if needed)
			{
				Config: `
resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.3.0"
  bgp_asn         = "65001"

  tags = [
    {
      key   = "Name"
      value = "Terraform-Test-Client-GatewayUpdated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "public_ip", "192.0.3.0"),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "bgp_asn", "65001"),
					resource.TestCheckResourceAttr("numspot_client_gateway.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_client_gateway.test", "tags.*", map[string]string{
						"key":   "Name",
						"value": "Terraform-Test-Client-GatewayUpdated",
					}),
					resource.TestCheckResourceAttrWith("numspot_client_gateway.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						return nil
					}),
				),
			},
		},
	})
}
