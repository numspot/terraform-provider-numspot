//go:build acc

package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccSecurityGroupResource_SingleInboundRule(t *testing.T) {
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
				Config: testSecurityGroupConfig_SingleInboundRule(netIpRange, name, descrition),
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
				Config: testSecurityGroupConfig_SingleInboundRule(netIpRange, name, descrition),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testSecurityGroupConfig_SingleInboundRule(netIpRange, name, description string) string {
	return fmt.Sprintf(`
resource "numspot_net" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  net_id      = numspot_net.net.id
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
}`, netIpRange, name, description)
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
resource "numspot_net" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  net_id      = numspot_net.net.id
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
resource "numspot_net" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  net_id      = numspot_net.net.id
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
resource "numspot_net" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  net_id      = numspot_net.net.id
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

func TestAccSecurityGroupResource_MultipleRules(t *testing.T) {
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
				Config: testSecurityGroupConfig_MultipleRules(netIpRange, name, descrition),
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
				Config: testSecurityGroupConfig_MultipleRules(netIpRange, name, descrition),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testSecurityGroupConfig_MultipleRules(netIpRange, name, description string) string {
	return fmt.Sprintf(`
resource "numspot_net" "net" {
  ip_range = %[1]q
}

resource "numspot_security_group" "test" {
  net_id      = numspot_net.net.id
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
      from_port_range = 443
      to_port_range   = 443
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]

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
