package test

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

const spaceID = "bba8c1df-609f-4775-9638-952d488502e6"

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataServiceAccount struct {
	name        string
	permissions []string
	roles       []string
}

// Generate checks to validate that resource 'numspot_service_account.test' has input data values
func getFieldMatchChecksServiceAccount(data StepDataServiceAccount) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_service_account.test", "name", data.name),
		resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", strconv.Itoa(len(data.roles))),
		resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", strconv.Itoa(len(data.permissions))),
	}

	for _, role := range data.roles {
		checks = append(
			checks,
			resource.TestCheckTypeSetElemAttrPair(
				"numspot_service_account.test",
				"roles.*", "data.numspot_roles."+strings.Replace(role, " ", "_", -1),
				"items.0.id",
			), // If field is a slice of ids
		)
	}

	for _, permission := range data.permissions {
		name := strings.Split(permission, ".")
		serviceName := name[0]
		resourceName := name[1]
		actionName := name[2]
		permName := generatePermissionName(serviceName, resourceName, actionName)
		checks = append(
			checks,
			resource.TestCheckTypeSetElemAttrPair("numspot_service_account.test", "global_permissions.*", "data.numspot_permissions."+permName, "items.0.id"), // If field is a slice of ids
		)
	}

	return checks
}

// Generate checks to validate that resource 'numspot_service_account.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksServiceAccount(dependenciesSuffix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{}
}

func TestAccServiceAccountResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	name := "My Service Account"
	nameUpdated := "My New Service Account"

	permissions := []string{
		"postgresql.backup.delete",
		"postgresql.backup.get",
	}
	permissionsUpdated_1 := permissions[:len(permissions)-1]
	permissionsUpdated_2 := []string{}

	roles := []string{
		"Postgres Admin",
		"Postgres Viewer",
	}
	rolesUpdated_1 := roles[:len(roles)-1]
	rolesUpdated_2 := []string{}
	rolesUpdated_3 := []string{
		"OCP Viewer",
		"OCP Admin",
	}

	// resource fields that cannot be updated in-place (requires replace)
	// None

	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataServiceAccount{
		name:        name,
		permissions: permissions,
		roles:       roles,
	}
	createChecks := append(
		getFieldMatchChecksServiceAccount(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_service_account.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues_1 := StepDataServiceAccount{
		name:        nameUpdated,
		permissions: permissionsUpdated_1,
		roles:       rolesUpdated_1,
	}
	updateChecks_1 := append(
		getFieldMatchChecksServiceAccount(updatePlanValues_1),

		resource.TestCheckResourceAttrWith("numspot_service_account.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues_2 := StepDataServiceAccount{
		name:        nameUpdated,
		permissions: permissionsUpdated_2,
		roles:       rolesUpdated_2,
	}
	updateChecks_2 := append(
		getFieldMatchChecksServiceAccount(updatePlanValues_2),

		resource.TestCheckResourceAttrWith("numspot_service_account.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues_3 := StepDataServiceAccount{
		name:        nameUpdated,
		permissions: permissionsUpdated_2,
		roles:       rolesUpdated_3,
	}
	updateChecks_3 := append(
		getFieldMatchChecksServiceAccount(updatePlanValues_3),

		resource.TestCheckResourceAttrWith("numspot_service_account.test", "id", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)
	/////////////////////////////////////////////////////////////////////////////////////

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // Create testing
				Config: config(t, basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksServiceAccount(acctest.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_service_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: config(t, updatePlanValues_1),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_1,
					getDependencyChecksServiceAccount(acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing Without Replace (if needed)
			{
				Config: config(t, updatePlanValues_2),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_2,
					getDependencyChecksServiceAccount(acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing Without Replace (if needed)
			{
				Config: config(t, updatePlanValues_3),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_3,
					getDependencyChecksServiceAccount(acctest.BASE_SUFFIX),
				)...),
			},
		},
	})
}

func testServiceAccountConfig(data StepDataServiceAccount) (string, error) {
	permissionsDataSource, permissionsIDs, err := initPermissions(data.permissions)
	if err != nil {
		return "", err
	}
	rolesDataSource, rolesIDs := initRoles(data.roles)

	conf := fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id           = %[1]q
  name               = %[2]q
  global_permissions = [%[3]s]
  roles              = [%[4]s]
}
`, spaceID, data.name, strings.Join(permissionsIDs, ","), strings.Join(rolesIDs, ","),
	)
	conf += permissionsDataSource + "\n" + rolesDataSource
	return conf, nil
}

func initPermissions(permissions []string) (string, []string, error) {
	permissionsIDs := make([]string, len(permissions))
	var permissionDataSource string

	for idx, permissionName := range permissions {
		splitPermissionName := strings.Split(permissionName, ".")
		if len(splitPermissionName) != 3 {
			return "", nil, fmt.Errorf("permission name: %s is not valid", permissionName)
		}

		permissionService := splitPermissionName[0]
		permissionResource := splitPermissionName[1]
		permissionAction := splitPermissionName[2]

		datasourceName := generatePermissionName(permissionService, permissionResource, permissionAction)
		permissionsIDs[idx] = fmt.Sprintf("data.numspot_permissions.%s.items.0.id", datasourceName)

		permissionDataSource += fmt.Sprintf(`data "numspot_permissions" %[1]q {
  space_id = %[2]q
  action   = %[3]q
  service  = %[4]q
  resource = %[5]q
}
`, datasourceName, spaceID, permissionAction, permissionService, permissionResource)
	}

	return permissionDataSource, permissionsIDs, nil
}

func initRoles(roles []string) (string, []string) {
	rolesIDs := make([]string, len(roles))
	var rolesDataSource string
	for idx, roleName := range roles {
		roleSerialized := strings.Replace(roleName, " ", "_", -1)
		rolesIDs[idx] = fmt.Sprintf("data.numspot_roles.%s.items.0.id", roleSerialized)
		rolesDataSource += fmt.Sprintf(`
data "numspot_roles" %[1]q {
  space_id = %[2]q
  name     = %[3]q
}
`, roleSerialized, spaceID, roleName)
	}

	return rolesDataSource, rolesIDs
}

func generatePermissionName(service, resource, action string) string {
	return fmt.Sprintf("perm_%[1]s_%[2]s_%[3]s", service, resource, action)
}

func config(t *testing.T, param StepDataServiceAccount) string {
	c, err := testServiceAccountConfig(param)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
