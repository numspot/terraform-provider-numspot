//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccPublicIpResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: createPublicIPConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "public_ip", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_public_ip.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: linkPublicIPToVMConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "link_public_ip", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
				ExpectNonEmptyPlan: true,
			},
			// Update testing
			{
				Config: UnlinkPublicIPConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("numspot_public_ip.test", "link_public_ip"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func createPublicIPConfig() string {
	return `
resource "numspot_public_ip" "test" {}`
}

func linkPublicIPToVMConfig() string {
	return `
resource "numspot_image" "test" {
  name               = "terraform-generated-image-for-public-ip-test"
  source_image_id    = "ami-026ce760"
  source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_vm" "vm" {
  image_id = numspot_image.test.id
  vm_type  = "tinav6.c1r1p3"
}

resource "numspot_public_ip" "test" {
  vm_id = numspot_vm.vm.id
}`
}

func UnlinkPublicIPConfig() string {
	return `
resource "numspot_image" "test" {
  name               = "terraform-generated-image-for-public-ip-test"
  source_image_id    = "ami-026ce760"
  source_region_name = "cloudgouv-eu-west-1"
}

resource "numspot_public_ip" "test" {}
`
}

func TestAccPublicIpResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	tagKey := "name"
	tagValue := "Terraform-Test-Public-Ip"
	tagValueUpdated := "Terraform-Test-Public-Ip-Updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: PublicIPConfig_Tags(tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_public_ip.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: PublicIPConfig_Tags(tagKey, tagValueUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.0.value", tagValueUpdated),
					resource.TestCheckResourceAttr("numspot_public_ip.test", "tags.#", "1"),
				),
			},
		},
	})
}

func PublicIPConfig_Tags(key, value string) string {
	return fmt.Sprintf(`
resource "numspot_public_ip" "test" {
  tags = [
    {
      key   = %[1]q
      value = %[2]q
    }
  ]
}
`, key, value)
}
