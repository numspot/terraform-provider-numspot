//go:build acc

package permission_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

func TestAccPermissionsDatasource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	spaceID := "bba8c1df-609f-4775-9638-952d488502e6"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDatasourceConfig(spaceID),
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
	pr := provider.TestAccProtoV6ProviderFactories

	spaceID := "bba8c1df-609f-4775-9638-952d488502e6"
	action := "get"
	service := "postgresql"
	resourceName := "backup"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDatasourceConfig_WithFilter(spaceID, action, service, resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_permissions.testdata", "items.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_permissions.testdata", "space_id", spaceID),
					provider.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_permissions.testdata", "items.*", map[string]string{
						"action":   action,
						"service":  service,
						"resource": resourceName,
					}),
				),
			},
		},
	})
}

func testAccPermissionsDatasourceConfig(spaceID string) string {
	return fmt.Sprintf(`
data "numspot_permissions" "testdata" {
  space_id = %[1]q
}`, spaceID)
}

func testAccPermissionsDatasourceConfig_WithFilter(spaceID, action, service, resource string) string {
	return fmt.Sprintf(`
data "numspot_permissions" "testdata" {
  space_id = %[1]q
  action   = %[2]q
  service  = %[3]q
  resource = %[4]q
}`, spaceID, action, service, resource)
}
