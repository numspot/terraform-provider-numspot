package route_table_test

import (
	"fmt"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestRouteTableResourceCreate(t *testing.T) {
	firstVPCID := "vpc-c3726ca8"
	secondVPCID := "vpc-802f94aa"

	var initialRouteTableID string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testRouteTableConfigCreate(firstVPCID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.main", "vpc_id", firstVPCID),
					resource.TestCheckResourceAttrSet("numspot_route_table.main", "id"),
					resource.TestCheckResourceAttrWith("numspot_route_table.main", "id", func(id string) error {
						initialRouteTableID = id
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_route_table.main", "link_route_tables.#", "0"),
					resource.TestCheckResourceAttr("numspot_route_table.main", "route_propagating_virtual_gateways.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_route_table.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testRouteTableConfigCreate(secondVPCID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.main", "vpc_id", secondVPCID),
					resource.TestCheckResourceAttrWith("numspot_route_table.main", "id", func(newRouteTableID string) error {
						require.NotEqual(t, initialRouteTableID, newRouteTableID)
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_route_table.main", "link_route_tables.#", "0"),
					resource.TestCheckResourceAttr("numspot_route_table.main", "route_propagating_virtual_gateways.#", "0"),
				),
			},
		},
	})
}

func testRouteTableConfigCreate(virtualPrivateCloudId string) string {
	return fmt.Sprintf(`
resource "numspot_route_table" "main" {
	vpc_id = %[1]q
}
`, virtualPrivateCloudId)
}
