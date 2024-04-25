//go:build acc

package provider

import (
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
			},
			// Update testing
			{
				Config: UnlinkPublicIPConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("numspot_public_ip.test", "link_public_ip"),
				),
			},
		},
	})
}

func createPublicIPConfig() string {
	return `resource "numspot_public_ip" "test" {}`
}

func linkPublicIPToVMConfig() string {
	return `
resource "numspot_vm" "vm" {
  image_id = "ami-060e019f"
  vm_type  = "tinav6.c1r1p3"
}

resource "numspot_public_ip" "test" {
  vm_id = numspot_vm.vm.id
}`
}

func UnlinkPublicIPConfig() string {
	return `resource "numspot_public_ip" "test" {}`
}
