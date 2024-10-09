package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccDHCPOptionsResource(t *testing.T) {
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
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 2 - ImportState testing
			{
				ResourceName:            "numspot_dhcp_options.terraform-dhcp-options-acctest",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Step 3 - Replace DHCP Options domain name
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name.replaced"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name.replaced"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 4 - Update DHCP Options tags
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name = "domain.name.replaced"
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest-update"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name.replaced"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest-update",
					}),
				),
			},
			// Step 5 - Delete DHCP Options with domain name
			{
				Config: ` `,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Step 6 - Create DHCP Options with domain name servers
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name_servers = ["192.0.0.1", "192.0.0.2"]
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 7 - Replace DHCP Options with domain name servers
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  domain_name_servers = ["192.0.0.3", "192.0.0.4"]
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest-replaced", "domain_name_servers", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 8 - Delete DHCP Options with domain name servers
			{
				Config: ` `,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Step 9 - Create DHCP Options with log servers
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  log_servers = ["192.0.0.1", "192.0.0.2"]
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "log_servers", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 10 - Replace DHCP Options with log servers
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  log_servers = ["192.0.0.3", "192.0.0.4"]
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest-replaced", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 11 - Delete DHCP Options with log servers
			{
				Config: ` `,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Step 12 - Create DHCP Options with NTP servers
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  ntp_servers = ["192.0.0.1", "192.0.0.2"]
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 13 - Replace DHCP Options with NTP servers
			{
				Config: `
resource "numspot_dhcp_options" "terraform-dhcp-options-acctest" {
  ntp_servers = ["192.0.0.3", "192.0.0.4"]
  tags = [
    {
      key   = "name"
      value = "terraform-dhcp-options-acctest"
    }
  ]
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest-replaced", "domain_name", "domain.name"),
					resource.TestCheckResourceAttr("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_dhcp_options.terraform-dhcp-options-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-dhcp-options-acctest",
					}),
				),
			},
			// Step 14 - Delete DHCP Options with NTP servers
			{
				Config: ` `,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}
