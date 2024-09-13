package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVmsDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	image_id := "ami-0b7df82c"
	vm_type := "ns-cus6-2c4r"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVmConfig(image_id, vm_type),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vms.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vms.testdata", "items.*", map[string]string{
						"id":       acctest.PAIR_PREFIX + "numspot_vm.test.id",
						"type":     vm_type,
						"image_id": image_id,
					}),
				),
				// ExpectNonEmptyPlan: true,
			},
		},
	})
}

func fetchVmConfig(image_id, vm_type string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test" {
  image_id  = %[1]q
  type      = %[2]q
  subnet_id = numspot_subnet.subnet.id
}

data "numspot_vms" "testdata" {
  ids        = [numspot_vm.test.id]
  depends_on = [numspot_vm.test]
}
`, image_id, vm_type)
}
