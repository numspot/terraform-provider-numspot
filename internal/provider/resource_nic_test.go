//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNicResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNicConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("numspot_nic.test", "sub", "value"),
				//resource.TestCheckResourceAttrWith("numspot_nic.test", "field", func(v string) error {
				//	require.NotEmpty(t, v)
				//	return nil
				//}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			//// Update testing
			//{
			//	Config: testNicConfig(),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("numspot_nic.test", "field", "value"),
			//		resource.TestCheckResourceAttrWith("numspot_nic.test", "field", func(v string) error {
			//			return nil
			//		}),
			//	),
			//},
		},
	})
}

func testNicConfig() string {
	return `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet.id
}`
}

func TestAccNicResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	tagKey := "name"
	tagValue := "Terraform-Test-Volume"
	tagValueUpdate := "Terraform-Test-Volume-Update"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNicConfig_Tags(tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			//// Update testing
			{
				Config: testNicConfig_Tags(tagKey, tagValueUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.0.value", tagValueUpdate),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testNicConfig_Tags(key, value string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet.id
  tags = [
	{
	  key 		= %[1]q
	  value	 	= %[2]q
	}
  ]
}`, key, value)
}
