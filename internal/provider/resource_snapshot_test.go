package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSnapshotResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	volumeId := "vol-toto"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSnapshotConfig(volumeId),
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
				ImportStateVerifyIgnore: []string{},
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

func testSnapshotConfig(volumeId string) string {
	return fmt.Sprintf(`
resource "numspot_snapshot" "test" {
	volume_id = %[1]q
}`, volumeId)
}
