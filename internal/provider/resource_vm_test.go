package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccVmResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	imageId := "ami-00b0c39a"
	vmType := "t2.small"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig_Create(imageId, vmType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.test", "image_id", imageId),
					resource.TestCheckResourceAttrWith("numspot_vm.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vm.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testVmConfig_Create(imageId, vmType),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}
func testVmConfig_Create(imageId, vmType string) string {
	return fmt.Sprintf(`
resource "numspot_vm" "test" {
	image_id = %[1]q
	vm_type = %[2]q
}
`, imageId, vmType)
}
func testVmConfig_Update() string {
	return `resource "numspot_vm" "test" {
    			}`
}
