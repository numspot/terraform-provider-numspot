//go:build acc

package vpcpeering_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccVpcPeeringsDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVpcPeeringConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_vpc_peerings.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_vpc_peerings.testdata", "items.*", map[string]string{
						"id": provider.PAIR_PREFIX + "numspot_vpc_peering.test.id",
					}),
				),
			},
		},
	})
}

func fetchVpcPeeringConfig() string {
	return `
resource "numspot_vpc" "source" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "accepter" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter.id
  source_vpc_id   = numspot_vpc.source.id
}

data "numspot_vpc_peerings" "testdata" {
  ids = [numspot_vpc_peering.test.id]
}
`
}
