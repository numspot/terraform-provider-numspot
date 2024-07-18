//go:build acc

package vm_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccVmsDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	image_id := "ami-0b7df82c"
	vm_type := "ns-cus6-2c4r"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVmConfig(image_id, vm_type),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vms.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vms.testdata", "items.*", map[string]string{
						"id":       provider.PAIR_PREFIX + "numspot_vm.test.id",
						"vm_type":  vm_type,
						"image_id": image_id,
					}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func fetchVmConfig(image_id, vm_type string) string {
	return fmt.Sprintf(`
resource "numspot_vm" "test" {
  image_id = %[1]q
  type     = %[2]q
}

data "numspot_vms" "testdata" {
  ids        = [numspot_vm.test.id]
  depends_on = [numspot_vm.test]
}
`, image_id, vm_type)
}
