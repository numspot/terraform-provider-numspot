package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVmsDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVmConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vms.testdata", "items.#", "1"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func fetchVmConfig() string {
	return `
resource "numspot_vm" "test" {
  image_id = "ami-026ce760"
  vm_type  = "ns-cus6-2c4r"
}

data "numspot_vms" "testdata" {
  ids        = [numspot_vm.test.id]
  depends_on = [numspot_vm.test]
}
`
}
