package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccInternetGatewaysDatasource(t *testing.T) {
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
				Config: `
resource "numspot_vpc" "terraform-dep-vpc-itw" {
  ip_range = "10.101.0.0/16"
  tags = [{
    key   = "name"
    value = "terraform-itw-acctest"
  }]
}

resource "numspot_internet_gateway" "terraform-itw-acctest" {
  vpc_id = numspot_vpc.terraform-dep-vpc-itw.id
  tags = [{
    key   = "name"
    value = "terraform-itw-acctest"
  }]
}

data "numspot_internet_gateways" "datasource-itw-acctest" {
  ids = [numspot_internet_gateway.terraform-itw-acctest.id]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_internet_gateways.datasource-itw-acctest", "items.#", "1"),
				),
			},
		},
	})
}
