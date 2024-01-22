package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccRouteTableResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	// Required
	netId := "vpc-d02edcb7"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testRouteTableConfig(netId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "net_id", netId),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_route_table.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testRouteTableConfig(netId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_route_table.test", "net_id", netId),
					resource.TestCheckResourceAttrWith("numspot_route_table.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
		},
	})
}

func testRouteTableConfig(netId string) string {
	return fmt.Sprintf(`
resource "numspot_route_table" "test" {
	net_id = %[1]q
}`, netId)
}
