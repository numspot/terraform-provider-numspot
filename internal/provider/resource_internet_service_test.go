package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccInternetServiceResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	netID := "vpc-087d645a"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testInternetServiceConfig_Create(netID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_internet_service.test", "net_id", netID),
					resource.TestCheckResourceAttrWith("numspot_internet_service.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_internet_service.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}
func testInternetServiceConfig_Create(netID string) string {
	return fmt.Sprintf(`resource "numspot_internet_service" "test" {
			net_id="%s"
  			}`, netID)
}
