package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccServiceAccountResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create testing
				Config: `
data "numspot_permissions" "perm_postgresql_backup_delete" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  action   = "delete"
  service  = "postgresql"
  resource = "backup"
}

data "numspot_permissions" "perm_postgresql_backup_get" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  action   = "get"
  service  = "postgresql"
  resource = "backup"
}

data "numspot_roles" "Postgres_Admin" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "Postgres Admin"
}

data "numspot_roles" "Postgres_Viewer" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "Postgres Viewer"
}

resource "numspot_service_account" "test" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "My Service Account"
  global_permissions = [
    data.numspot_permissions.perm_postgresql_backup_delete.items.0.id,
    data.numspot_permissions.perm_postgresql_backup_get.items.0.id,
  ]
  roles = [
    data.numspot_roles.Postgres_Admin.items.0.id,
    data.numspot_roles.Postgres_Viewer.items.0.id,
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", "My Service Account"),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", "2"),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_service_account.test",
						"roles.*",
						"data.numspot_roles.Postgres_Admin",
						"items.0.id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_service_account.test",
						"roles.*",
						"data.numspot_roles.Postgres_Viewer",
						"items.0.id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_service_account.test",
						"global_permissions.*",
						"data.numspot_permissions.perm_postgresql_backup_get",
						"items.0.id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_service_account.test",
						"global_permissions.*",
						"data.numspot_permissions.perm_postgresql_backup_delete",
						"items.0.id",
					),
					resource.TestCheckResourceAttrWith("numspot_service_account.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_service_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			// 3 - Update testing Without Replace
			{
				Config: `
data "numspot_permissions" "perm_postgresql_backup_delete" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  action   = "delete"
  service  = "postgresql"
  resource = "backup"
}

data "numspot_roles" "Postgres_Admin" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "Postgres Admin"
}

resource "numspot_service_account" "test" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "My Service Account Updated"
  global_permissions = [
    data.numspot_permissions.perm_postgresql_backup_delete.items.0.id,
  ]
  roles = [
    data.numspot_roles.Postgres_Admin.items.0.id,
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", "My Service Account Updated"),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_service_account.test",
						"roles.*",
						"data.numspot_roles.Postgres_Admin",
						"items.0.id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_service_account.test",
						"global_permissions.*",
						"data.numspot_permissions.perm_postgresql_backup_delete",
						"items.0.id",
					),
					resource.TestCheckResourceAttrWith("numspot_service_account.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
			},
			// 4 - Update testing Without Replace
			{
				Config: `
resource "numspot_service_account" "test" {
  space_id           = "bba8c1df-609f-4775-9638-952d488502e6"
  name               = "My Service Account Updated"
  global_permissions = []
  roles              = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", "My Service Account Updated"),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", "0"),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", "0"),
					resource.TestCheckResourceAttrWith("numspot_service_account.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
			},
			// 5 - Update testing Without Replace (if needed)
			{
				Config: `
data "numspot_roles" "OCP_Viewer" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "OCP Viewer"
}

data "numspot_roles" "OCP_Admin" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "OCP Admin"
}
resource "numspot_service_account" "test" {
  space_id           = "bba8c1df-609f-4775-9638-952d488502e6"
  name               = "My Service Account Updated"
  global_permissions = []
  roles = [
    data.numspot_roles.OCP_Viewer.items.0.id,
    data.numspot_roles.OCP_Admin.items.0.id,
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_service_account.test", "name", "My Service Account Updated"),
					resource.TestCheckResourceAttr("numspot_service_account.test", "roles.#", "2"),
					resource.TestCheckResourceAttr("numspot_service_account.test", "global_permissions.#", "0"),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_service_account.test",
						"roles.*",
						"data.numspot_roles.OCP_Viewer",
						"items.0.id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"numspot_service_account.test",
						"roles.*",
						"data.numspot_roles.OCP_Admin",
						"items.0.id",
					),
					resource.TestCheckResourceAttrWith("numspot_service_account.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
			},
		},
	})
}
