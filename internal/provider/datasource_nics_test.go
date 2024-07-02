package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNicsDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchNicConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_nics.testdata", "items.#", "1"),
					resource.TestCheckResourceAttrPair("data.numspot_nics.testdata", "items.0.id", "numspot_nic.test", "id"),
				),
			},
		},
	})
}

func fetchNicConfig() string {
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
}

data "numspot_nics" "testdata" {
  ids        = [numspot_nic.test.id]
  depends_on = [numspot_nic.test]
}
`
}
