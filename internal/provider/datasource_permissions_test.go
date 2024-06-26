//go:build acc

package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccPermissionsDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	spaceID := "68134f98-205b-4de4-8b85-f6a786ef6481"
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
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	spaceID := "68134f98-205b-4de4-8b85-f6a786ef6481"
	action := "get"
	service := "network"
	resourceName := "link"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDatasourceConfig_WithFilter(spaceID, action, service, resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_permissions.testdata", "items.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_permissions.testdata", "items.0.action", action),
					resource.TestCheckResourceAttr("data.numspot_permissions.testdata", "items.0.service", service),
					resource.TestCheckResourceAttr("data.numspot_permissions.testdata", "items.0.resource", resourceName),
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
