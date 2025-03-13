package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// 1 - Create Snapshot from Volume
// 2 - Import
// 3 - Update attributes (Snapshot from Volume)
// 4 - Update attributes with Replace (Snapshot from Volume)
// 5 - Recreate (Snapshot from Volume)
//
// 6 - Create Snapshot from Snapshot
// 7 - Import
// 8 - Update attributes (Snapshot from Snapshot)
// 9 - Update attributes with Replace (Snapshot from Snapshot)
// 10 - Recreate (Snapshot from Snapshot)

func TestAccSnapshotResource(t *testing.T) {
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
			{ // 1 - Create Snapshot from Volume
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "A beautiful snapshot"
  tags = [
    {
      key   = "name"
      value = "Snapshot-Terraform-Test"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", "A beautiful snapshot"),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Snapshot-Terraform-Test",
					}),
					resource.TestCheckResourceAttrPair("numspot_snapshot.test", "volume_id", "numspot_volume.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			{ // 2 - ImportState testing
				ResourceName:            "numspot_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{ // 3 - Update attributes (Snapshot from Volume)
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "A beautiful snapshot"
  tags = [
    {
      key   = "name"
      value = "Snapshot-Terraform-Test-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", "A beautiful snapshot"),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Snapshot-Terraform-Test-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_snapshot.test", "volume_id", "numspot_volume.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 4 - Update attributes with Replace (Snapshot from Volume)
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "A beautiful snapshot but updated"
  tags = [
    {
      key   = "name"
      value = "Snapshot-Terraform-Test-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", "A beautiful snapshot but updated"),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Snapshot-Terraform-Test-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_snapshot.test", "volume_id", "numspot_volume.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 5 - Recreate (Snapshot from Volume)
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test_recreated" {
  volume_id   = numspot_volume.test.id
  description = "A beautiful snapshot"
  tags = [
    {
      key   = "name"
      value = "Snapshot-Terraform-Test"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test_recreated", "description", "A beautiful snapshot"),
					resource.TestCheckResourceAttr("numspot_snapshot.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Snapshot-Terraform-Test",
					}),
					resource.TestCheckResourceAttrPair("numspot_snapshot.test_recreated", "volume_id", "numspot_volume.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test_recreated", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 6 - Create Snapshot from Snapshot
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "snapshot" {
  volume_id = numspot_volume.test.id
}

resource "numspot_snapshot" "test" {
  source_snapshot_id = numspot_snapshot.snapshot.id
  source_region_name = "cloudgouv-eu-west-1"
  description        = "A beautiful snapshot"
  tags = [
    {
      key   = "name"
      value = "Snapshot-Terraform-Test"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", "A beautiful snapshot"),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Snapshot-Terraform-Test",
					}),
					resource.TestCheckResourceAttrPair("numspot_snapshot.test", "source_snapshot_id", "numspot_snapshot.snapshot", "id"),

					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			{ // 7 - ImportState testing
				ResourceName:            "numspot_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_region_name", "source_snapshot_id"},
			},
			{ // 8 - Update attributes (Snapshot from Snapshot)
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "snapshot" {
  volume_id = numspot_volume.test.id
}

resource "numspot_snapshot" "test" {
  source_snapshot_id = numspot_snapshot.snapshot.id
  source_region_name = "cloudgouv-eu-west-1"
  description        = "A beautiful snapshot"
  tags = [
    {
      key   = "name"
      value = "Snapshot-Terraform-Test-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", "A beautiful snapshot"),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Snapshot-Terraform-Test-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_snapshot.test", "source_snapshot_id", "numspot_snapshot.snapshot", "id"),

					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			{ // 9 - Update attributes with Replace (Snapshot from Snapshot)
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "snapshot" {
  volume_id = numspot_volume.test.id
}

resource "numspot_snapshot" "test" {
  source_snapshot_id = numspot_snapshot.snapshot.id
  source_region_name = "cloudgouv-eu-west-1"
  description        = "A beautiful snapshot but updated"
  tags = [
    {
      key   = "name"
      value = "Snapshot-Terraform-Test-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test", "description", "A beautiful snapshot but updated"),
					resource.TestCheckResourceAttr("numspot_snapshot.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Snapshot-Terraform-Test-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_snapshot.test", "source_snapshot_id", "numspot_snapshot.snapshot", "id"),

					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			{ // 10 - Recreate (Snapshot from Snapshot)
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "snapshot" {
  volume_id = numspot_volume.test.id
}

resource "numspot_snapshot" "test_recreated" {
  source_snapshot_id = numspot_snapshot.snapshot.id
  source_region_name = "cloudgouv-eu-west-1"
  description        = "A beautiful snapshot but updated"
  tags = [
    {
      key   = "name"
      value = "Snapshot-Terraform-Test-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_snapshot.test_recreated", "description", "A beautiful snapshot but updated"),
					resource.TestCheckResourceAttr("numspot_snapshot.test_recreated", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_snapshot.test_recreated", "tags.*", map[string]string{
						"key":   "name",
						"value": "Snapshot-Terraform-Test-Updated",
					}),
					resource.TestCheckResourceAttrPair("numspot_snapshot.test_recreated", "source_snapshot_id", "numspot_snapshot.snapshot", "id"),

					resource.TestCheckResourceAttrWith("numspot_snapshot.test_recreated", "id", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
