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
	spaceID := "bba8c1df-609f-4775-9638-952d488502e6"
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
	spaceID := "bba8c1df-609f-4775-9638-952d488502e6"
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

func TestAccServiceAccountResource_Roles(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	spaceID := "bba8c1df-609f-4775-9638-952d488502e6"
	name := "My Service Account"
	updatedName := "My New Service Account"

	roles := []string{
		"8d9706cc-c77a-499c-bb67-3597644f6d27",
		"fd4c0997-157a-42ba-89ac-241e54c05124",
	}

	updatedRoles := roles[:len(roles)-1]
	updatedRoles2 := []string{}
	updatedRoles3 := []string{
		"3a3afcfb-555c-4495-943a-70973940cc17",
		"44541946-1310-49c0-b43a-e1c64aa3925c",
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testServiceAccountConfig_Roles(spaceID, name, roles),
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
				Config: testServiceAccountConfig_Roles(spaceID, updatedName, updatedRoles),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", strconv.Itoa(len(updatedRoles))),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.0", updatedRoles[0]),
				),
			},
			// Update testing
			{
				Config: testServiceAccountConfig_Roles(spaceID, updatedName, updatedRoles2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", strconv.Itoa(len(updatedRoles2))),
				),
			},
			// Update testing
			{
				Config: testServiceAccountConfig_Roles(spaceID, updatedName, updatedRoles3),
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

func testServiceAccountConfig_Roles(spaceID, name string, roles []string) string {
	var rolesList string

	if len(roles) > 0 {
		rolesList = fmt.Sprintf(`["%s"]`, strings.Join(roles, `", "`))
	} else {
		rolesList = fmt.Sprintf(`[]`)
	}

	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = %[2]q

  roles = %[3]s
}`, spaceID, name, rolesList)
}
