//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccSpaceDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	organisationId := "67d97ad4-3005-48dc-a392-60a97ab5097c"
	name := "the space"
	description := "the description"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchSpaceConfig(organisationId, name, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.numspot_space.testdata", "id", func(id string) error {
						require.NotEmpty(t, id)
						return nil
					}),
					resource.TestCheckResourceAttr("data.numspot_space.testdata", "organisation_id", organisationId),
					resource.TestCheckResourceAttr("data.numspot_space.testdata", "name", name),
					resource.TestCheckResourceAttr("data.numspot_space.testdata", "description", description),
				),
			},
		},
	})
}

func fetchSpaceConfig(organisationId string, name string, description string) string {
	return fmt.Sprintf(`
resource "numspot_space" "test" {
  organisation_id = %[1]q
  name            = %[2]q
  description     = %[3]q
}

data "numspot_space" "testdata" {
  space_id   = numspot_space.test.id
  depends_on = [numspot_space.test]
}
	`, organisationId, name, description)
}
