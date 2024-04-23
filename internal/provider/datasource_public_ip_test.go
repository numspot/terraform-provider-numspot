//go:build acc

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPublicIpsDatasource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchPublicIpsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_public_ip.testdata", "public_ips.#", "1"),
				),
			},
		},
	})
}

func fetchPublicIpsConfig() string {
	return `
resource "numspot_public_ip" "test" {}

data "numspot_public_ip" "testdata" {
  ids        = [numspot_public_ip.test.id]
  depends_on = [numspot_public_ip.test]
}


`
}
