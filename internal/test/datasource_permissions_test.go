package test

import (
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccPermissionsDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: `
data "numspot_permissions" "testdata" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.numspot_permissions.testdata", "items.#", func(v string) error {
						count, err := strconv.Atoi(v)
						if err != nil {
							return err
						}

						require.Greater(t, count, 0)
						return nil
					}),
				),
			},
		},
	})
}

func TestAccPermissionsDatasource_WithFilter(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: `
data "numspot_permissions" "testdata" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  action   = "get"
  service  = "postgresql"
  resource = "backup"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_permissions.testdata", "items.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_permissions.testdata", "space_id", "bba8c1df-609f-4775-9638-952d488502e6"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_permissions.testdata", "items.*", map[string]string{
						"action":   "get",
						"service":  "postgresql",
						"resource": "backup",
					}),
				),
			},
		},
	})
}
