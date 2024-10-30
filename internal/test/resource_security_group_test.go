package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccSecurityGroupResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// 1 - Create Security group
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test.id
  name        = "security-group-name"
  description = "security-group-description"
  inbound_rules = [
    {
      from_port_range = 453
      to_port_range   = 453
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
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
    }
  ]
  outbound_rules = [
    {
      from_port_range = 455
      to_port_range   = 455
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 90
      to_port_range   = 90
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.#", "3"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "outbound_rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "453",
						"to_port_range":   "453",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "80",
						"to_port_range":   "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "22",
						"to_port_range":   "22",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "455",
						"to_port_range":   "455",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "90",
						"to_port_range":   "90",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test", "id"),
				),
			},
			// 2 - Import
			{
				ResourceName:            "numspot_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test.id
  name        = "security-group-name"
  description = "security-group-description"
  inbound_rules = [
    {
      from_port_range = 453
      to_port_range   = 453
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 20
      to_port_range   = 20
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
  outbound_rules = []
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.#", "2"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "outbound_rules.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "453",
						"to_port_range":   "453",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "20",
						"to_port_range":   "20",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test", "id"),
				),
			},
			// 4 - Update testing Without Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id        = numspot_vpc.test.id
  name          = "security-group-name"
  description   = "security-group-description"
  inbound_rules = []
  outbound_rules = [
    {
      from_port_range = 455
      to_port_range   = 455
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 90
      to_port_range   = 90
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 70
      to_port_range   = 70
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.#", "0"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "outbound_rules.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "455",
						"to_port_range":   "455",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "90",
						"to_port_range":   "90",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "80",
						"to_port_range":   "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "70",
						"to_port_range":   "70",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test", "id"),
				),
			},
			// 5 - Update testing With Replace
			{
				Config: `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test.id
  name        = "security-group-name-updated"
  description = "security-group-description-updated"
  inbound_rules = [
    {
      from_port_range = 453
      to_port_range   = 453
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
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
    }
  ]
  outbound_rules = [
    {
      from_port_range = 455
      to_port_range   = 455
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 90
      to_port_range   = 90
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name-updated"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description-updated"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.#", "3"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "outbound_rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "453",
						"to_port_range":   "453",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "80",
						"to_port_range":   "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "22",
						"to_port_range":   "22",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "455",
						"to_port_range":   "455",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "90",
						"to_port_range":   "90",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test", "id"),
				),
			},
			// <== If resource has required dependencies ==>
			// 6 - Update testing With Replace of dependency resource and with Replace of the resource
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "test_new" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.test_new.id
  name        = "security-group-name-updated"
  description = "security-group-description-updated"
  inbound_rules = [
    {
      from_port_range = 453
      to_port_range   = 453
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
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
    }
  ]
  outbound_rules = [
    {
      from_port_range = 455
      to_port_range   = 455
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 90
      to_port_range   = 90
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test", "name", "security-group-name-updated"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "description", "security-group-description-updated"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test", "inbound_rules.#", "3"),
					resource.TestCheckResourceAttr("numspot_security_group.test", "outbound_rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "453",
						"to_port_range":   "453",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "80",
						"to_port_range":   "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "inbound_rules.*", map[string]string{
						"from_port_range": "22",
						"to_port_range":   "22",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "455",
						"to_port_range":   "455",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test", "outbound_rules.*", map[string]string{
						"from_port_range": "90",
						"to_port_range":   "90",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test", "vpc_id", "numspot_vpc.test_new", "id"),
				),
			},
			// 7- recreate testing
			{
				Config: `
resource "numspot_vpc" "test_recreate" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test_recreate" {
  vpc_id      = numspot_vpc.test_recreate.id
  name        = "security-group-name"
  description = "security-group"
  inbound_rules = [
    {
      from_port_range = 453
      to_port_range   = 453
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
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
    }
  ]
  outbound_rules = [
    {
      from_port_range = 455
      to_port_range   = 455
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 90
      to_port_range   = 90
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "description", "security-group"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "inbound_rules.#", "3"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "outbound_rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "inbound_rules.*", map[string]string{
						"from_port_range": "453",
						"to_port_range":   "453",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "inbound_rules.*", map[string]string{
						"from_port_range": "80",
						"to_port_range":   "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "inbound_rules.*", map[string]string{
						"from_port_range": "22",
						"to_port_range":   "22",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "outbound_rules.*", map[string]string{
						"from_port_range": "455",
						"to_port_range":   "455",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "outbound_rules.*", map[string]string{
						"from_port_range": "90",
						"to_port_range":   "90",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test_recreate", "vpc_id", "numspot_vpc.test_recreate", "id"),
				),
			},
			// 8- reset rules
			{
				Config: `
resource "numspot_vpc" "test_recreate" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test_recreate" {
  vpc_id         = numspot_vpc.test_recreate.id
  name           = "security-group-name"
  description    = "security-group"
  inbound_rules  = []
  outbound_rules = []
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "description", "security-group"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "inbound_rules.#", "0"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "outbound_rules.#", "0"),
					resource.TestCheckResourceAttrPair("numspot_security_group.test_recreate", "vpc_id", "numspot_vpc.test_recreate", "id"),
				),
			},
			// 9- Add rules
			{
				Config: `
resource "numspot_vpc" "test_recreate" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test_recreate" {
  vpc_id      = numspot_vpc.test_recreate.id
  name        = "security-group-name"
  description = "security-group"
  inbound_rules = [
    {
      from_port_range = 453
      to_port_range   = 453
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
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
    }
  ]
  outbound_rules = [
    {
      from_port_range = 455
      to_port_range   = 455
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 90
      to_port_range   = 90
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
  ]
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-SecurityGroup"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "name", "security-group-name"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "description", "security-group"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-SecurityGroup",
					}),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "inbound_rules.#", "3"),
					resource.TestCheckResourceAttr("numspot_security_group.test_recreate", "outbound_rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "inbound_rules.*", map[string]string{
						"from_port_range": "453",
						"to_port_range":   "453",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "inbound_rules.*", map[string]string{
						"from_port_range": "80",
						"to_port_range":   "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "inbound_rules.*", map[string]string{
						"from_port_range": "22",
						"to_port_range":   "22",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "outbound_rules.*", map[string]string{
						"from_port_range": "455",
						"to_port_range":   "455",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_security_group.test_recreate", "outbound_rules.*", map[string]string{
						"from_port_range": "90",
						"to_port_range":   "90",
					}),
					resource.TestCheckResourceAttrPair("numspot_security_group.test_recreate", "vpc_id", "numspot_vpc.test_recreate", "id"),
				),
			},
		},
	})
}
