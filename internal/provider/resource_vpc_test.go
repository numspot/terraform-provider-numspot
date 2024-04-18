//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNetResource(t *testing.T) {
	t.Parallel()
	ipRange := "10.101.0.0/16"
	ipRangeUpdated := "10.102.0.0/16"

	previousId := ""

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetConfig(ipRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", ipRange),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						previousId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetConfig(ipRangeUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", ipRangeUpdated),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEqual(t, previousId, v)
						return nil
					}),
				),
			},
		},
	})
}

func testNetConfig(ipRange string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = %[1]q
}`, ipRange)
}

func TestAccNetResource_Tags(t *testing.T) {
	t.Parallel()
	//tagName := "Terraform Provider VPC"
	//updatedTagName := "Terraform Provider VPC - 2"

	previousId := ""

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetConfig_Tags(),
				//Config: testNetConfig_Tags(tagName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						previousId = v
						return nil
					}),
					//resource.TestCheckResourceAttr("numspot_vpc.test", "tags.0.key", "Name"),
					//resource.TestCheckResourceAttr("numspot_vpc.test", "tags.0.value", tagName),
					//resource.TestCheckResourceAttr("numspot_vpc.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetConfig_Tags(),
				//Config: testNetConfig_Tags(updatedTagName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.Equal(t, previousId, v)
						return nil
					}),
				),
			},
		},
	})
}

/*func testNetConfig_Tags(name string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
  tenancy  = "default"
  tags = [
    {
      key   = "Name"
      value = %[1]q
    }
  ]
}`, name)
}*/

func testNetConfig_Tags() string {
	return `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
  tenancy  = "default"
}`
}
