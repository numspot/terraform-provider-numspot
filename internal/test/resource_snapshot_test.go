package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

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
			//////// Test snapshot created from Volume
			{ // 1 - Create testing
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
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace
			{
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
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
			},
			// 4 - Update testing With Replace (if needed)
			{
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
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						return nil
					}),
				),
			},

			// <== If resource has required dependencies ==>
			{ // 5 - Reset the resource to initial state (resource tied to a subresource) in prevision of next test
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
			},
			// 6 - Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_volume" "test_new" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test_new.id
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
					resource.TestCheckResourceAttrPair("numspot_snapshot.test", "volume_id", "numspot_volume.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						return nil
					}),
				),
			},

			//////// Test snapshot created from another Snapshot
			{ // 7 - Create testing
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
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 8 - ImportState testing
			{
				ResourceName:            "numspot_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_region_name", "source_snapshot_id"},
			},
			// 9 - Update testing Without Replace (if needed)
			{
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
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
			},
			// 10 - Update testing With Replace
			{
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
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						resourceId = v
						return nil
					}),
				),
			},

			// <== If resource has required dependencies ==>
			{ // 11 - Reset the resource to initial state (resource tied to a subresource) in prevision of next test
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
			},

			// 12 - Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "snapshot_new" {
  volume_id = numspot_volume.test.id
}

resource "numspot_snapshot" "test" {
  source_snapshot_id = numspot_snapshot.snapshot_new.id
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
					resource.TestCheckResourceAttrPair("numspot_snapshot.test", "source_snapshot_id", "numspot_snapshot.snapshot_new", "id"),

					resource.TestCheckResourceAttrWith("numspot_snapshot.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						resourceId = v
						return nil
					}),
				),
			},
		},
	})
}
