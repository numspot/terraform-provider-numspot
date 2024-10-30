package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccSpaceResource(t *testing.T) {
	t.Skip()

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
resource "numspot_space" "test" {
  organisation_id = "67d97ad4-3005-48dc-a392-60a97ab5097c"
  name            = "quiet space"
  description     = "A quiet space"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_space.test", "name", "quiet space"),
					resource.TestCheckResourceAttr("numspot_space.test", "description", "A quiet space"),
					resource.TestCheckResourceAttrWith("numspot_space.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:      "numspot_space.test",
				ImportState:       true,
				ImportStateVerify: true,
			},

			// 3 - Update testing With Replace (if needed)
			{
				Config: `
resource "numspot_space" "test" {
  organisation_id = "67d97ad4-3005-48dc-a392-60a97ab5097c"
  name            = "quiet space updated"
  description     = "A new quiet space"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_space.test", "name", "quiet space updated"),
					resource.TestCheckResourceAttr("numspot_space.test", "description", "A new quiet space"),
					resource.TestCheckResourceAttrWith("numspot_space.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
