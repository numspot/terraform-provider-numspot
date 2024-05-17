//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceAccountResource_Basic(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	spaceID := "68134f98-205b-4de4-8b85-f6a786ef6481"
	name := "My Service Account"
	updatedName := "My New Service Account"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testServiceAccountConfig(spaceID, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", name),
					resource.TestCheckResourceAttr("numspot_service_account.test", "space_id", spaceID),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_service_account.test",
				ImportState:             true,
				ImportStateVerify:       false,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing
			{
				Config: testServiceAccountConfig(spaceID, updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
				),
			},
		},
	})
}

func testServiceAccountConfig(spaceID, name string) string {
	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = %[2]q
}`, spaceID, name)
}
