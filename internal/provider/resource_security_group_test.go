//go:build acc

package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccSecurityGroupResource_SingleInboundRule_WithReplace(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	netIpRange := "10.101.0.0/16"

	randName := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("security-group-name-%d", randName)
	description := fmt.Sprintf("security-group-description-%d", randName)

	nameUpdated := name + "_updated"
	descriptionUpdated := description + "_updated"

	var securityGroupId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig_SingleInboundRule(netIpRange, name, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", name),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", description),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						securityGroupId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSecurityGroupConfig_SingleInboundRule(netIpRange, nameUpdated, descriptionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						if v == securityGroupId {
							return errors.New("Id should be different after Update with replace")
						}
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", nameUpdated),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", descriptionUpdated),
				),
			},
		},
	})
}

func testSecurityGroupConfig_SingleInboundRule(netIpRange, name, description string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = %[2]q
  description = %[3]q
  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}
`, netIpRange, name, description)
}

func TestAccSecurityGroupResource_CoupleInboundRule(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	netIpRange := "10.101.0.0/16"

	randName := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("security-group-name-%d", randName)
	descrition := fmt.Sprintf("security-group-description-%d", randName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig_CoupleInboundRule(netIpRange, name, descrition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", name),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", descrition),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inbound_rules", "outbound_rules"},
			},
			// Update testing
			{
				Config: testSecurityGroupConfig_CoupleInboundRule(netIpRange, name, descrition),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testSecurityGroupConfig_CoupleInboundRule(netIpRange, name, description string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = %[2]q
  description = %[3]q
  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 22
      to_port_range   = 22
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
}`, netIpRange, name, description)
}

func TestAccSecurityGroupResource_SingleOutboundRule(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	netIpRange := "10.101.0.0/16"

	randName := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("security-group-name-%d", randName)
	descrition := fmt.Sprintf("security-group-description-%d", randName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig_SingleOutboundRule(netIpRange, name, descrition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", name),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", descrition),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSecurityGroupConfig_SingleOutboundRule(netIpRange, name, descrition),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testSecurityGroupConfig_SingleOutboundRule(netIpRange, name, description string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = %[2]q
  description = %[3]q
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}`, netIpRange, name, description)
}

func TestAccSecurityGroupResource_CoupleOutboundRule(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	netIpRange := "10.101.0.0/16"

	randName := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("security-group-name-%d", randName)
	descrition := fmt.Sprintf("security-group-description-%d", randName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig_CoupleOutboundRule(netIpRange, name, descrition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", name),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", descrition),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSecurityGroupConfig_CoupleOutboundRule(netIpRange, name, descrition),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testSecurityGroupConfig_CoupleOutboundRule(netIpRange, name, description string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = %[2]q
  description = %[3]q
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 443
      to_port_range   = 443
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}`, netIpRange, name, description)
}

func TestAccSecurityGroupResource_MultipleRules_NoReplace(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	netIpRange := "10.101.0.0/16"

	randName := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("security-group-name-%d", randName)
	description := fmt.Sprintf("security-group-description-%d", randName)

	rule1PortRange := "443"
	rule2PortRange := "80"

	rule1PortRangeUpdated := "453"
	rule2PortRangeUpdated := "90"

	var securityGroupId string
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig_MultipleRules_NoReplace(netIpRange, name, description, rule1PortRange, rule2PortRange),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", name),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", description),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						securityGroupId = v
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.0.from_port_range", rule1PortRange),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.1.from_port_range", rule2PortRange),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSecurityGroupConfig_MultipleRules_NoReplace(netIpRange, name, description, rule1PortRangeUpdated, rule2PortRangeUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						if v != securityGroupId {
							return errors.New("Id should be the same after Update without replace")
						}
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.0.from_port_range", rule1PortRangeUpdated),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.1.from_port_range", rule2PortRangeUpdated),
				),
			},
		},
	})
}

func testSecurityGroupConfig_MultipleRules_NoReplace(netIpRange, name, description, rule1PortRange, rule2PortRange string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = %[2]q
  description = %[3]q

  inbound_rules = [
    {
      from_port_range = %[4]s
      to_port_range   = %[4]s
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = %[5]s
      to_port_range   = %[5]s
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]

  outbound_rules = [
    {
      from_port_range = %[4]s
      to_port_range   = %[4]s
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = %[5]s
      to_port_range   = %[5]s
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}`, netIpRange, name, description, rule1PortRange, rule2PortRange)
}

func TestAccSecurityGroupResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	netIpRange := "10.101.0.0/16"

	randName := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("security-group-name-%d", randName)
	descrition := fmt.Sprintf("security-group-description-%d", randName)

	tagKey := "name"
	tagValue := "Terraform-Test-SecurityGroup"
	tagValueUpdate := "Terraform-Test-SecurityGroup-Updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfig_Tags(netIpRange, name, descrition, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", name),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", descrition),
					resource.TestCheckResourceAttrWith("numspot_security_group.test", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSecurityGroupConfig_Tags(netIpRange, name, descrition, tagKey, tagValueUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.0.value", tagValueUpdate),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testSecurityGroupConfig_Tags(netIpRange, name, description, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.net.id
  name        = %[2]q
  description = %[3]q
  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
}`, netIpRange, name, description, tagKey, tagValue)
}
