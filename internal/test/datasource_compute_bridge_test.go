//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccComputeBridgeDatasource(t *testing.T) {
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
resource "numspot_vpc" "vpc-source-compute-bridge-datasource-acctest" {
  ip_range = "10.101.1.0/24"
  tags = [{
    key   = "name"
    value = "terraform-compute-bridge-acctest"
  }]
}

resource "numspot_vpc" "vpc-destination-compute-bridge-datasource-acctest" {
  ip_range = "10.101.2.0/24"
  tags = [{
    key   = "name"
    value = "terraform-compute-bridge-acctest"
  }]
}

resource "numspot_compute_bridge" "resource-compute-bridge-datasource-acctest" {
  source_vpc_id      = numspot_vpc.vpc-source-compute-bridge-datasource-acctest.id
  destination_vpc_id = numspot_vpc.vpc-destination-compute-bridge-datasource-acctest.id
}

data "numspot_compute_bridges" "datasource-compute-bridge-acctest" {
  depends_on = [numspot_compute_bridge.resource-compute-bridge-datasource-acctest]
}

output "selected_bridge" {
  value = data.numspot_compute_bridges.datasource-compute-bridge-acctest
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_compute_bridges.datasource-compute-bridge-acctest", "items.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.numspot_compute_bridges.datasource-compute-bridge-acctest",
						"items.*",
						map[string]string{
							"source_ip_range":      "10.101.1.0/24",
							"destination_ip_range": "10.101.2.0/24",
						},
					),
				),
			},
		},
	})
}
