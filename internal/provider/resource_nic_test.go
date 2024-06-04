//go:build acc

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNicResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNicConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testNicConfig() string {
	return `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet.id
}`
}

func TestAccNicResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	tagKey := "name"
	tagValue := "Terraform-Test-Volume"
	tagValueUpdate := "Terraform-Test-Volume-Update"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNicConfig_Tags(tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			//// Update testing
			{
				Config: testNicConfig_Tags(tagKey, tagValueUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.0.value", tagValueUpdate),
					resource.TestCheckResourceAttr("numspot_nic.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testNicConfig_Tags(key, value string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet.id
  tags = [
    {
      key   = %[1]q
      value = %[2]q
    }
  ]
}`, key, value)
}

func TestAccNicResourceUpdateWithoutReplace(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	var nic_id, security_group_id string
	description := "The nic"
	descriptionUpdated := "The better nic"

	securityGroupSuffix := ""
	securityGroupSuffixUpdated := "_updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNicConfigUpdateWithoutReplace(description, securityGroupSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						nic_id = v
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", description),
					resource.TestCheckResourceAttr("numspot_nic.test", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "security_group_ids.0", func(v string) error {
						require.NotEmpty(t, v)
						security_group_id = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testNicConfigUpdateWithoutReplace(descriptionUpdated, securityGroupSuffixUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						if nic_id != v {
							return errors.New("Id should be the same after Update without replace")
						}
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_nic.test", "description", descriptionUpdated),
					resource.TestCheckResourceAttr("numspot_nic.test", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "security_group_ids.0", func(v string) error {
						require.NotEmpty(t, v)
						if security_group_id == v {
							return errors.New("security_group id should be different after update")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testNicConfigUpdateWithoutReplace(
	description string,
	security_group_suffix string,
) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "security_group" {
  net_id      = numspot_vpc.vpc.id
  name        = "security_group"
  description = "numspot_security_group description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_security_group" "security_group_updated" {
  net_id      = numspot_vpc.vpc.id
  name        = "security_group_updated"
  description = "numspot_security_group description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_nic" "test" {
  subnet_id          = numspot_subnet.subnet.id
  description        = %[1]q
  security_group_ids = [numspot_security_group.security_group%[2]s.id]
  depends_on         = [numspot_security_group.security_group%[2]s]
}`, description, security_group_suffix)
}

func TestAccNicResourceUpdateWithReplace(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	var nic_id, subnet_id string
	subnetSuffix := ""
	subnetSuffixUpdated := "_updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNicConfigUpdateWithReplace(subnetSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						nic_id = v
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "subnet_id", func(v string) error {
						require.NotEmpty(t, v)
						subnet_id = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testNicConfigUpdateWithReplace(subnetSuffixUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_nic.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						if nic_id == v {
							return errors.New("Id should be different after Update with replace")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "subnet_id", func(v string) error {
						require.NotEmpty(t, v)
						if subnet_id == v {
							return errors.New("Subnet Id should be different after Update")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testNicConfigUpdateWithReplace(
	subnet_suffix string,
) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_subnet" "subnet_updated" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet%[1]s.id
}`, subnet_suffix)
}
