package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccVolumesDatasource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	volumeType := "standard"
	volumeSize := 11
	volumeAZ := "eu-west-2a"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchVolumesConfigById(volumeType, volumeSize, volumeAZ),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_volumes.datasource_test", "volumes.#", "1"),
					resource.TestCheckResourceAttr("data.numspot_volumes.datasource_test", "volumes.0.id", "numspot_volume.test.id"),
					//resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdatabyid", "dhcp_options.#", "1"),
					//resource.TestCheckResourceAttr("data.numspot_dhcp_options.testdatabyid", "dhcp_options.0.domain_name", "foo.bar"),
				),
			},
		},
	})
}

func fetchVolumesConfigById(volumeType string, volumeSize int, volumeAZ string) string {
	return fmt.Sprintf(`
resource "numspot_volume" "test" {
	type 					= %[1]q
	size 					= %[2]d
	availability_zone_name 	= %[3]q
}

data "numspot_volumes" "datasource_test" {
	ids = [numspot_volume.test.id]
}
`, volumeType, volumeSize, volumeAZ)

}
