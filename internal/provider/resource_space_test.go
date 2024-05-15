//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccSpaceResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	organisationID := "67d97ad4-3005-48dc-a392-60a97ab5097c"
	spaceName := "10.101.0.0/16"
	spaceDescription := "10.101.1.0/24"
	updatedSpaceName := "10.101.2.0/24"
	updatedSpaceDescription := "10.101.2.0/24"

	// Computed
	spaceId := ""

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSpaceConfig(organisationID, spaceName, spaceDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_space.test", "name", spaceName),
					resource.TestCheckResourceAttr("numspot_space.test", "description", spaceDescription),
					resource.TestCheckResourceAttrWith("numspot_space.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						spaceId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_space.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSpaceConfig(organisationID, updatedSpaceName, updatedSpaceDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_space.test", "name", updatedSpaceName),
					resource.TestCheckResourceAttr("numspot_space.test", "description", updatedSpaceDescription),
					resource.TestCheckResourceAttrWith("numspot_space.test", "id", func(v string) error {
						require.NotEqual(t, v, spaceId)
						return nil
					}),
				),
			},
		},
	})
}

func testSpaceConfig(organisationID, spaceName, spaceDescription string) string {
	return fmt.Sprintf(`
resource "numspot_space" "test" {
  organisation_id = %[1]q
  name            = %[2]q
  description     = %[3]q
}`, organisationID, spaceName, spaceDescription)
}
