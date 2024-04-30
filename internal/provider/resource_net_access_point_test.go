//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNetAccessPointResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetAccessPointConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_net_access_point.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_net_access_point.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetAccessPointConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_net_access_point.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testNetAccessPointConfig_Create() string {
	return `resource "numspot_net_access_point" "test" {}`
}

func testNetAccessPointConfig_Update() string {
	return `resource "numspot_net_access_point" "test" {
}`
}

func TestAccNetAccessPointResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	tagKey := "name"
	tagValue := "Terraform-Test-Vpc-Access-Point"
	tagValueUpdate := "Terraform-Test-Vpc-Access-Point-Update"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetAccessPointConfig_Tags(tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_net_access_point.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetAccessPointConfig_Tags(tagKey, tagValueUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "tags.0.value", tagValueUpdate),
					resource.TestCheckResourceAttr("numspot_net_access_point.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testNetAccessPointConfig_Tags(tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/24"
}

resource "numspot_net_access_point" "test" {
  net_id       = numspot_vpc.test.id
  service_name = "com.outscale.cloudgouv-eu-west-1.oos"

  tags = [
    {
      key   = %[1]q
      value = %[2]q
    }
  ]
}
`, tagKey, tagValue)
}
