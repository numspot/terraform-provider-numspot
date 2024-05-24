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
					resource.TestCheckResourceAttr("data.numspot_public_ips.public_ip_data_source", "items.#", "1"),
				),
			},
		},
	})
}

func fetchPublicIpsConfig() string {
	return `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "internet_gateway" {
  vpc_id     = numspot_vpc.vpc.id
  depends_on = [numspot_nic.nic]
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "nic" {
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_public_ip" "public_ip" {
  nic_id     = numspot_nic.nic.id
  depends_on = [numspot_nic.nic, numspot_internet_gateway.internet_gateway]
}

data "numspot_public_ips" "public_ip_data_source" {
  ids        = [numspot_public_ip.public_ip.id]
  depends_on = [numspot_public_ip.public_ip]
}
`
}
