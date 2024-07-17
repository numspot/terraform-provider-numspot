//go:build acc

package provider

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/utils_acctest"
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
		resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", strconv.Itoa(len(data.roles))),
	}

	for _, role := range data.roles {
		checks = append(checks, resource.TestCheckTypeSetElemAttr("numspot_service_account.test", "roles.*", role))
	}

	for _, permission := range data.permissions {
		checks = append(checks, resource.TestCheckTypeSetElemAttr("numspot_service_account.test", "global_permissions.*", permission))
	}

	return checks
}

// Generate checks to validate that resource 'numspot_service_account.test' is properly linked to given subresources
// If resource has no dependencies, return empty array
func getDependencyChecksServiceAccount(dependenciesPrefix string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{}
}

func TestAccServiceAccountResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place
	name := "My Service Account"
	nameUpdated := "My New Service Account"

	// TODO : use permissions/roles datasource instead of hardcoded UUID
	permissions := []string{
		"0288e52e-3853-49d0-b1d1-2daef11be8ab",
		"33fe3c63-4a3b-4e46-a28d-c24d7d2d64de",
	}
	permissionsUpdated_1 := permissions[:len(permissions)-1]
	permissionsUpdated_2 := []string{}

	roles := []string{
		"6f275ef2-6651-4150-86e7-9e6a51fa1f56",
		"92061163-c643-481f-96e2-37788864fd78",
	}
	rolesUpdated_1 := roles[:len(roles)-1]
	rolesUpdated_2 := []string{}
	rolesUpdated_3 := []string{
		"31b75fdd-319d-4517-b543-80e18b88a410",
		"708696a9-cb19-4d57-861e-ba5212d6b933",
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
				Config: testServiceAccountConfig(basePlanValues),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					createChecks,
					getDependencyChecksServiceAccount(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_service_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Update testing Without Replace (if needed)
			{
				Config: testServiceAccountConfig(updatePlanValues_1),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_1,
					getDependencyChecksServiceAccount(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing Without Replace (if needed)
			{
				Config: testServiceAccountConfig(updatePlanValues_2),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_2,
					getDependencyChecksServiceAccount(utils_acctest.BASE_SUFFIX),
				)...),
			},
			// Update testing Without Replace (if needed)
			{
				Config: testServiceAccountConfig(updatePlanValues_3),
				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
					updateChecks_3,
					getDependencyChecksServiceAccount(utils_acctest.BASE_SUFFIX),
				)...),
			},
		},
	})
}

func testServiceAccountConfig(data StepDataServiceAccount) string {
	permissions := utils_acctest.ListToStringList(data.permissions)
	roles := utils_acctest.ListToStringList(data.roles)

	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id           = %[1]q
  name               = %[2]q
  global_permissions = %[3]s
  roles              = %[4]s

}`, spaceID, data.name, permissions, roles)
}
