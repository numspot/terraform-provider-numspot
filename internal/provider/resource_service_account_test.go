//go:build acc

package provider

import (
	"fmt"
	"strconv"
	"strings"
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

func TestAccServiceAccountResource_GlobalPermission(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	spaceID := "68134f98-205b-4de4-8b85-f6a786ef6481"
	name := "My Service Account"
	updatedName := "My New Service Account"

	permissions := []string{
		"94034915-045e-4196-a7e7-714aa207db68",
		"766b2dca-4238-4d39-a5ea-86b99318f2b5",
	}

	updatedPermissions1 := permissions[:len(permissions)-1]
	updatedPermissions2 := []string{}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testServiceAccountConfig_GlobalPermission(spaceID, name, permissions),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", name),
					resource.TestCheckResourceAttr("numspot_service_account.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", strconv.Itoa(len(permissions))),
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
				Config: testServiceAccountConfig_GlobalPermission(spaceID, updatedName, updatedPermissions1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", strconv.Itoa(len(updatedPermissions1))),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.0", updatedPermissions1[0]),
				),
			},
			// Update testing
			{
				Config: testServiceAccountConfig_GlobalPermission(spaceID, updatedName, updatedPermissions2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", strconv.Itoa(len(updatedPermissions2))),
				),
			},
		},
	})
}

func testServiceAccountConfig_GlobalPermission(spaceID, name string, permissions []string) string {
	var permissionsList string

	if len(permissions) > 0 {
		permissionsList = fmt.Sprintf(`["%s"]`, strings.Join(permissions, `", "`))
	} else {
		permissionsList = fmt.Sprintf(`[]`)
	}

	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = %[2]q

  global_permissions = %[3]s
}`, spaceID, name, permissionsList)
}
