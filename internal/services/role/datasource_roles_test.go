//go:build acc

package role_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccRolesDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	spaceID := "68134f98-205b-4de4-8b85-f6a786ef6481"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDatasourceConfig(spaceID),
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
	pr := provider.TestAccProtoV6ProviderFactories

	spaceID := "68134f98-205b-4de4-8b85-f6a786ef6481"
	name := "kubernetes Viewer"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDatasourceConfig_WithFilter(spaceID, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_roles.testdata", "items.#", "1"),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_roles.testdata", "items.*", map[string]string{
						"name":     name,
						"space_id": spaceID,
					}),
				),
			},
		},
	})
}

func testAccRolesDatasourceConfig(spaceID string) string {
	return fmt.Sprintf(`
data "numspot_roles" "testdata" {
  space_id = %[1]q
}`, spaceID)
}

func testAccRolesDatasourceConfig_WithFilter(spaceID, name string) string {
	return fmt.Sprintf(`
data "numspot_roles" "testdata" {
  space_id = %[1]q
  name     = %[2]q
}`, spaceID, name)
}
