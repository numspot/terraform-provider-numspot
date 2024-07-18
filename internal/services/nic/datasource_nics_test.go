//go:build acc

package nic_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccNicsDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchNicConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_nics.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_nics.testdata", "items.*", map[string]string{
						"id":        provider.PAIR_PREFIX + "numspot_nic.test.id",
						"subnet_id": provider.PAIR_PREFIX + "numspot_subnet.subnet.id",
					}),
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
  ids = [numspot_nic.test.id]
}
`
}
