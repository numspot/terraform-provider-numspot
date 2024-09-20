package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccSpaceDatasource(t *testing.T) {
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
resource "numspot_space" "test" {
  organisation_id = "67d97ad4-3005-48dc-a392-60a97ab5097c"
  name            = "the space"
  description     = "the description"
}

data "numspot_space" "testdata" {
  space_id        = numspot_space.test.space_id
  organisation_id = numspot_space.test.organisation_id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.numspot_space.testdata", "id", "numspot_space.test", "space_id"),
					resource.TestCheckResourceAttr("data.numspot_space.testdata", "organisation_id", "67d97ad4-3005-48dc-a392-60a97ab5097c"),
					resource.TestCheckResourceAttr("data.numspot_space.testdata", "name", "the space"),
					resource.TestCheckResourceAttr("data.numspot_space.testdata", "description", "the description"),
				),
			},
		},
	})
}
