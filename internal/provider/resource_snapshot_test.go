package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccSnapshotResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSnapshotConfig_Create(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
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
			{
				Config: testSnapshotConfig_Update(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testSnapshotConfig_Create() string {
	return fmt.Sprintf(`resource "numspot_snapshot" "test" {
  			}`)
}

func testSnapshotConfig_Update() string {
	return `resource "numspot_snapshot" "test" {
    			}`
}
