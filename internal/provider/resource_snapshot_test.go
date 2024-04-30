//go:build acc

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSnapshotResourceFromVolume(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	description := "A beautiful snapshot"
	updated_description := "An even more beautiful snapshot"

	var snapshot_id string
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSnapshotFromVolumeConfig(description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", description),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						snapshot_id = v
						return nil
					}),
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
			{
				Config: testSnapshotFromVolumeConfig(updated_description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", updated_description),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						if snapshot_id == v {
							return errors.New("Id should be different after Update")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testSnapshotFromVolumeConfig(description string) string {
	return fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = %[1]q
}`, description)
}

func TestAccSnapshotResourceFromSnapshot(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	description := "A beautiful snapshot"
	updated_description := "An even more beautiful snapshot"

	var snapshot_id string
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSnapshotFromSnapshotConfig(description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", description),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						snapshot_id = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"progress", "state", "source_region_name", "source_snapshot_id"},
			},
			// Update testing
			{
				Config: testSnapshotFromSnapshotConfig(updated_description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", updated_description),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						if v == "" {
							return errors.New("Id should not be empty")
						}
						if snapshot_id == v {
							return errors.New("Id should be different after Update")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testSnapshotFromSnapshotConfig(description string) string {
	return fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2a"

}

resource "numspot_snapshot" "snapshot_from_volume" {
  volume_id   = numspot_volume.test.id
  description = %[1]q
}

resource "numspot_snapshot" "test" {
  source_snapshot_id = numspot_snapshot.snapshot_from_volume.id
  source_region_name = "eu-west-2"
}`, description)
}

func TestAccSnapshotResourceFromVolume_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	description := "A beautiful snapshot"
	tagKey := "name"
	tagValue := "Snapshot-Terraform-Test"
	tagValueUpdate := "Snapshot-Terraform-Test-Updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSnapshotFromVolumeConfig_Tags(description, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", description),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
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
			{
				Config: testSnapshotFromVolumeConfig_Tags(description, tagKey, tagValueUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", description),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.0.value", tagValueUpdate),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testSnapshotFromVolumeConfig_Tags(description, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = %[1]q
  tags = [
    {
      key 	= %[2]q
      value = %[3]q
    }
  ]
}`, description, tagKey, tagValue)
}
