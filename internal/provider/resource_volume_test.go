//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumeResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	volumeType := "standard"
	volumeSize := 11
	updatedVolumeSize := 22
	volumeAZ := "eu-west-2a"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: createVolumeConfig(volumeType, volumeSize, volumeAZ),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "type", volumeType),
					// resource.TestCheckResourceAttr("numspot_volume.test", "size", fmt.Sprintf("%d", volumeSize)),
					resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", volumeAZ),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_volume.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			// Update testing
			{
				Config: createVolumeConfig(volumeType, updatedVolumeSize, volumeAZ),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("numspot_volume.test", "field", "value"),
				//resource.TestCheckResourceAttrWith("numspot_volume.test", "field", func(v string) error {
				//	return nil
				//}),
				),
			},
		},
	})
}

func createVolumeConfig(volumeType string, volumeSize int, volumeAZ string) string {
	t := fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = %[1]q
  size                   = %[2]d
  availability_zone_name = %[3]q
}`, volumeType, volumeSize, volumeAZ)
	return t
}

func TestAccVolumeResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	volumeType := "standard"
	volumeSize := 11
	updatedVolumeSize := 22
	volumeAZ := "cloudgouv-eu-west-1a"
	tagKey := "name"
	tagValue := "Terraform-Test-Volume"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: createVolumeConfig_Tags(volumeType, volumeSize, volumeAZ, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "type", volumeType),
					resource.TestCheckResourceAttr("numspot_volume.test", "size", fmt.Sprintf("%d", volumeSize)),
					resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", volumeAZ),
					resource.TestCheckResourceAttr("numspot_volume.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_volume.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_volume.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_volume.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			// Update testing
			{
				Config: createVolumeConfig(volumeType, updatedVolumeSize, volumeAZ),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_volume.test", "type", volumeType),
					resource.TestCheckResourceAttr("numspot_volume.test", "size", fmt.Sprintf("%d", updatedVolumeSize)),
					resource.TestCheckResourceAttr("numspot_volume.test", "availability_zone_name", volumeAZ),
					resource.TestCheckResourceAttr("numspot_volume.test", "tags.#", "0"),
				),
			},
		},
	})
}

func createVolumeConfig_Tags(volumeType string, volumeSize int, volumeAZ, tagKey, tagValue string) string {
	t := fmt.Sprintf(`
resource "numspot_volume" "test" {
  type                   = %[1]q
  size                   = %[2]d
  availability_zone_name = %[3]q
  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
}`, volumeType, volumeSize, volumeAZ, tagKey, tagValue)
	return t
}
