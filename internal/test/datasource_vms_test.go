package test

import (
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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}

data "numspot_vms" "testdata" {
  ids = [numspot_vm.test.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vms.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vms.testdata", "items.*", map[string]string{
						"id":       acctest.PairPrefix + "numspot_vm.test.id",
						"type":     "ns-cus6-2c4r",
						"image_id": "ami-0b7df82c",
					}),
				),
			},
		},
	})
}
