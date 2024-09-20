package test

import (
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccRolesDatasource(t *testing.T) {
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
data "numspot_roles" "testdata" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.numspot_roles.testdata", "items.#", func(v string) error {
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

func TestAccRolesDatasource_WithFilter(t *testing.T) {
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
data "numspot_roles" "testdata" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "kubernetes Viewer"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_roles.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_roles.testdata", "items.*", map[string]string{
						"name": "kubernetes Viewer",
					}),
				),
			},
		},
	})
}
