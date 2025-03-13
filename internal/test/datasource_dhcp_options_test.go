package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccDHCPOptionsDataSource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Step 1 - Create DHCP Options with domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name"
  tags = [{
    key   = "name"
    value = "terraform-dhcp-options-acctest"
  }]
}

data "numspot_dhcp_options" "datasource-dhcp-options-acctest" {
  ids = [numspot_dhcp_options.terraform-dhcp-options-acctest.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_dhcp_options.datasource-dhcp-options-acctest", "items.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("data.numspot_dhcp_options.datasource-dhcp-options-acctest", "items.*", map[string]string{
						"domain_name": "domain.name",
					}),
				),
			},
		},
	})
}
