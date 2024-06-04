//go:build acc

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNetResource(t *testing.T) {
	t.Parallel()
	ipRange := "10.101.0.0/16"
	ipRangeUpdated := "10.102.0.0/16"

	previousId := ""

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetConfig(ipRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", ipRange),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						previousId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetConfig(ipRangeUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", ipRangeUpdated),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEqual(t, previousId, v)
						return nil
					}),
				),
			},
		},
	})
}

func testNetConfig(ipRange string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = %[1]q
}`, ipRange)
}

func TestAccNetResourceUpdateWithReplace(t *testing.T) {
	t.Parallel()

	domainName := "foo.bar"
	domainNameUpdated := "foo.bar.updated"
	ipRange := "10.101.0.0/16"
	ipRangeUpdated := "10.102.0.0/16"

	previousId := ""

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetConfigUpdateWithReplace(domainName, ipRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", ipRange),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "dhcp_options_set_id", func(v string) error {
						require.NotEmpty(t, v)
						previousId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetConfigUpdateWithReplace(domainNameUpdated, ipRangeUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.test", "ip_range", ipRangeUpdated),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEqual(t, previousId, v)
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "dhcp_options_set_id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
		},
	})
}

func testNetConfigUpdateWithReplace(domainName, ipRange string) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "test" {
  domain_name = %[1]q
}

resource "numspot_vpc" "test" {
  ip_range            = %[2]q
  dhcp_options_set_id = numspot_dhcp_options.test.id
}`, domainName, ipRange)
}

func TestAccNetResource_Tags(t *testing.T) {
	t.Parallel()
	tagValue := "Terraform Provider VPC"
	updatedTagValue := "Terraform Provider VPC - 2"

	previousId := ""

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetConfig_Tags(tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						previousId = v
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.0.key", "Name"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNetConfig_Tags(updatedTagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.Equal(t, previousId, v)
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.0.key", "Name"),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.0.value", updatedTagValue),
					resource.TestCheckResourceAttr("numspot_vpc.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testNetConfig_Tags(name string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
  tenancy  = "default"
  tags = [
    {
      key   = "Name"
      value = %[1]q
    }
  ]
}`, name)
}

func TestAccVpcResourceUpdateWithoutReplace(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	var vpc_id, dhcp_option_id string
	dhcpOptionSuffix := ""
	dhcpOptionSuffixUpdated := "_updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testVpcConfigUpdateWithoutReplace(dhcpOptionSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						vpc_id = v
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "dhcp_options_set_id", func(v string) error {
						require.NotEmpty(t, v)
						dhcp_option_id = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_vpc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testVpcConfigUpdateWithoutReplace(dhcpOptionSuffixUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						if vpc_id != v {
							return errors.New("Id should be the same after Update without replace")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_vpc.test", "dhcp_options_set_id", func(v string) error {
						require.NotEmpty(t, v)
						if dhcp_option_id == v {
							return errors.New("dhcp_options_set_id should be different after update")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testVpcConfigUpdateWithoutReplace(
	dhcp_option_suffix string,
) string {
	return fmt.Sprintf(`
resource "numspot_dhcp_options" "test" {
  domain_name = "domain"
}

resource "numspot_dhcp_options" "test_updated" {
  domain_name = "domain"
}
resource "numspot_vpc" "test" {
  ip_range            = "10.101.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.test%[1]s.id
}`, dhcp_option_suffix)
}
