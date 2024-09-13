package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccPublicIpsDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fetchPublicIpsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_public_ips.testdata", "items.#", "1"),
					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_public_ips.testdata", "items.*", map[string]string{
						"id":        acctest.PAIR_PREFIX + "numspot_public_ip.test.id",
						"public_ip": acctest.PAIR_PREFIX + "numspot_public_ip.test.public_ip",
					}),
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

resource "numspot_public_ip" "test" {
  nic_id     = numspot_nic.nic.id
  depends_on = [numspot_nic.nic, numspot_internet_gateway.internet_gateway]
}

data "numspot_public_ips" "testdata" {
  ids = [numspot_public_ip.test.id]
}
`
}
