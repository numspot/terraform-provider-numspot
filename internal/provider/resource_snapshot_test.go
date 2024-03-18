package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSnapshotResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSnapshotConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("numspot_snapshot.test", "field", "value"),
				//resource.TestCheckResourceAttrWith("numspot_snapshot.test", "field", func(v string) error {
				//	require.NotEmpty(t, v)
				//	return nil
				//}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"progress", "state"},
			},
			// Update testing
			//{
			//	Config: testSnapshotConfig(),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("numspot_snapshot.test", "field", "value"),
			//		resource.TestCheckResourceAttrWith("numspot_snapshot.test", "field", func(v string) error {
			//			return nil
			//		}),
			//	),
			//},
		},
	})
}

func testSnapshotConfig() string {
	return `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2a"
}

resource "numspot_snapshot" "test" {
  volume_id = numspot_volume.test.id
}`
}
