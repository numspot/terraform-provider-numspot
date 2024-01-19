package provider

import (
  "fmt"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/stretchr/testify/require"
)

func TestAccVmResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_vm.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_vm.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testVmConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_vm.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}
func testVmConfig_Create() string {
	return fmt.Sprintf(`resource "numspot_vm" "test" {
  			}`)
}
func testVmConfig_Update() string {
		return `resource "numspot_vm" "test" {
    			}`
}
