///go:build acc

package provider

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const spaceID = "bba8c1df-609f-4775-9638-952d488502e6"

func TestAccServiceAccountResource_Basic(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	name := "My Service Account"
	updatedName := "My New Service Account"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testServiceAccountConfig(name),
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
				Config: testServiceAccountConfig(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
				),
			},
		},
	})
}

func testServiceAccountConfig(name string) string {
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
	name := "My Service Account"
	updatedName := "My New Service Account"

	permissions := []string{
		"0288e52e-3853-49d0-b1d1-2daef11be8ab",
		"33fe3c63-4a3b-4e46-a28d-c24d7d2d64de",
	}

	updatedPermissions1 := permissions[:len(permissions)-1]
	updatedPermissions2 := []string{}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testServiceAccountConfig_GlobalPermission(name, permissions),
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
				Config: testServiceAccountConfig_GlobalPermission(updatedName, updatedPermissions1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", strconv.Itoa(len(updatedPermissions1))),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.0", updatedPermissions1[0]),
				),
			},
			// Update testing
			{
				Config: testServiceAccountConfig_GlobalPermission(updatedName, updatedPermissions2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", strconv.Itoa(len(updatedPermissions2))),
				),
			},
		},
	})
}

func testServiceAccountConfig_GlobalPermission(name string, permissions []string) string {
	var permissionsList string

	if len(permissions) > 0 {
		permissionsList = fmt.Sprintf(`["%s"]`, strings.Join(permissions, `", "`))
	} else {
		permissionsList = `[]`
	}

	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = %[2]q

  global_permissions = %[3]s
}`, spaceID, name, permissionsList)
}

func TestAccServiceAccountResource_Roles(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	name := "My Service Account"
	updatedName := "My New Service Account"

	roles := []string{
		"6f275ef2-6651-4150-86e7-9e6a51fa1f56",
		"92061163-c643-481f-96e2-37788864fd78",
	}

	updatedRoles := roles[:len(roles)-1]
	updatedRoles2 := []string{}
	updatedRoles3 := []string{
		"31b75fdd-319d-4517-b543-80e18b88a410",
		"708696a9-cb19-4d57-861e-ba5212d6b933",
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testServiceAccountConfig_Roles(name, roles),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", name),
					resource.TestCheckResourceAttr("numspot_service_account.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", strconv.Itoa(len(roles))),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.0", roles[0]),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.1", roles[1]),
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
				Config: testServiceAccountConfig_Roles(updatedName, updatedRoles),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", strconv.Itoa(len(updatedRoles))),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.0", updatedRoles[0]),
				),
			},
			// Update testing
			{
				Config: testServiceAccountConfig_Roles(updatedName, updatedRoles2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", strconv.Itoa(len(updatedRoles2))),
				),
			},
			// Update testing
			{
				Config: testServiceAccountConfig_Roles(updatedName, updatedRoles3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", strconv.Itoa(len(updatedRoles3))),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.0", updatedRoles3[0]),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.1", updatedRoles3[1]),
				),
			},
		},
	})
}

func testServiceAccountConfig_Roles(name string, roles []string) string {
	var rolesList string

	if len(roles) > 0 {
		rolesList = fmt.Sprintf(`["%s"]`, strings.Join(roles, `", "`))
	} else {
		rolesList = `[]`
	}

	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = %[2]q

  roles = %[3]s
}`, spaceID, name, rolesList)
}
