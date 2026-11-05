//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccComputeBridgeResource(t *testing.T) {
	acct := acctest.NewAccTest(t)
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
resource "numspot_vpc" "vpc-source" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "vpc-dest" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_compute_bridge" "resource-compute-bridge-resource-acctest" {
  source_vpc_id      = numspot_vpc.vpc-source.id
  destination_vpc_id = numspot_vpc.vpc-dest.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"numspot_compute_bridge.resource-compute-bridge-resource-acctest",
						"source_vpc_id",
						"numspot_vpc.vpc-source",
						"id",
					),
					resource.TestCheckResourceAttrPair(
						"numspot_compute_bridge.resource-compute-bridge-resource-acctest",
						"destination_vpc_id",
						"numspot_vpc.vpc-dest",
						"id",
					),
				),
			},
		},
	})
}
