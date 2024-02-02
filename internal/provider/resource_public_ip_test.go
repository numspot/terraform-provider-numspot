package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccPublicIpResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	vmid := "i-93372752" //labeled test_tf_publicIP in OSC cockpit

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testPublicIpConfig(),
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
				Config: testPublicIpConfig_Update(vmid),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "link_public_ip", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// Update testing
			{
				Config: testPublicIpConfig_UpdateUnlink(),
				/*Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_public_ip.test", "link_public_ip", func(v string) error {
						require.Empty(t, v)
						return nil
					}),
				),*/
			},
		},
	})
}

func testPublicIpConfig() string {
	return fmt.Sprintf(`resource "numspot_public_ip" "test" {

}`)
}

func testPublicIpConfig_Update(vmid string) string {
	return fmt.Sprintf(`resource "numspot_public_ip" "test" {
                        vm_id="%s"
                        }`, vmid)
}

func testPublicIpConfig_UpdateUnlink() string {
	return fmt.Sprintf(`resource "numspot_public_ip" "test" {
                        }`)
}
