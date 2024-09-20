package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccDhcpOptionsResource(t *testing.T) {
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
resource "numspot_dhcp_options" "test" {
  domain_name = "foo.bar"
  tags = [
    {
      key   = "name"
      value = "Terraform Provider DHCP Options"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "domain_name", "foo.bar"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform Provider DHCP Options",
					}),
					resource.TestCheckResourceAttrWith("numspot_dhcp_options.test", "id", func(v string) error {
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
				ResourceName:            "numspot_dhcp_options.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_dhcp_options" "test" {
  domain_name = "foo.bar"
  tags = [
    {
      key   = "name"
      value = "Terraform Provider DHCP Options Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "domain_name", "foo.bar"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform Provider DHCP Options Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_dhcp_options.test", "id", func(v string) error {
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
resource "numspot_dhcp_options" "test" {
  domain_name = "bar.foo"
  tags = [
    {
      key   = "name"
      value = "Terraform Provider DHCP Options Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "domain_name", "bar.foo"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform Provider DHCP Options Updated",
					}),
					resource.TestCheckResourceAttrWith("numspot_dhcp_options.test", "id", func(v string) error {
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
