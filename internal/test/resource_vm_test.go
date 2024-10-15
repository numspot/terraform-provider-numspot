package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccVmResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	//	vmDependencies := `
	//resource "numspot_vpc" "terraform-dep-vm-vpc" {
	//  ip_range = "10.101.0.0/16"
	//  tags = [{
	//    key   = "name"
	//    value = "terraform-dep-vm-vpc"
	//  }]
	//}
	//
	//resource "numspot_subnet" "terraform-dep-vm-subnet" {
	//  vpc_id                 = numspot_vpc.terraform-dep-vm-vpc.id
	//  ip_range               = "10.101.1.0/24"
	//  availability_zone_name = "cloudgouv-eu-west-1a"
	//  tags = [{
	//    key   = "name"
	//    value = "terraform-dep-vm-subnet"
	//  }]
	//}
	//`

	vmUpdateDependencies := `
resource "numspot_vpc" "terraform-dep-vm-vpc" {
  ip_range = "10.101.0.0/16"
  tags = [{
    key   = "name"
    value = "terraform-dep-vm-vpc"
  }]
}

resource "numspot_subnet" "terraform-dep-vm-subnet" {
  vpc_id                 = numspot_vpc.terraform-dep-vm-vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [{
    key   = "name"
    value = "terraform-dep-vm-subnet"
  }]
}

resource "numspot_keypair" "terraform-dep-vm-keypair" {
  name       = "keypair-name"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEm78d7vfikcOXDdvT0yioYUDm3spxjVws/xnL0J5f0P"
}

resource "numspot_security_group" "terraform-dep-vm-sg" {
  vpc_id      = numspot_vpc.terraform-dep-vm-vpc.id
  name        = "terraform-dep-vm-sg-name"
  description = "terraform-dep-vm-sg-description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			//			// Step 1 - Create VM
			//			{
			//				Config: vmDependencies + `
			//resource "numspot_vm" "numspot-vm-acctest" {
			//  subnet_id = numspot_subnet.terraform-dep-vm-subnet.id
			//  image_id  = "ami-00669acb"
			//  type      = "ns-cus6-2c4r"
			//
			//  tags = [{
			//    key   = "name"
			//    value = "terraform-vm-acctest"
			//  }]
			//}`,
			//				Check: resource.ComposeAggregateTestCheckFunc(
			//					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "subnet_id", "numspot_subnet.terraform-dep-vm-subnet", "id"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "image_id", "ami-00669acb"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "type", "ns-cus6-2c4r"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "tags.#", "1"),
			//					resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.numspot-vm-acctest", "tags.*", map[string]string{
			//						"key":   "name",
			//						"value": "terraform-vm-acctest",
			//					}),
			//				),
			//			},
			//			// Step 2 - Import
			//			{
			//				ResourceName:            "numspot_vm.numspot-vm-acctest",
			//				ImportState:             true,
			//				ImportStateVerify:       true,
			//				ImportStateVerifyIgnore: []string{"id"},
			//			},
			//			// Step 3 - Update VM attributes
			//			{
			//				Config: vmUpdateDependencies + `
			//resource "numspot_vm" "numspot-vm-acctest" {
			//  subnet_id          = numspot_subnet.terraform-dep-vm-subnet.id
			//  keypair_name       = numspot_keypair.terraform-dep-vm-keypair.name
			//  security_group_ids = [numspot_security_group.terraform-dep-vm-sg.id]
			//
			//  image_id                    = "ami-00669acb"
			//  type                        = "ns-eco6-2c2r"
			//  user_data                   = "dXNlci1kYXRhLWVuY29kZWQ="
			//  initiated_shutdown_behavior = "terminate"
			//
			//  tags = [{
			//    key   = "name"
			//    value = "terraform-vm-acctest-update"
			//  }]
			//}
			//`,
			//				Check: resource.ComposeAggregateTestCheckFunc(
			//					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "subnet_id", "numspot_subnet.terraform-dep-vm-subnet", "id"),
			//					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "keypair_name", "numspot_keypair.terraform-dep-vm-keypair", "name"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "security_group_ids.#", "1"),
			//					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "security_group_ids.0", "numspot_security_group.terraform-dep-vm-sg", "id"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "image_id", "ami-00669acb"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "type", "ns-eco6-2c2r"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "user_data", "dXNlci1kYXRhLWVuY29kZWQ="),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "initiated_shutdown_behavior", "terminate"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "tags.#", "1"),
			//					resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.numspot-vm-acctest", "tags.*", map[string]string{
			//						"key":   "name",
			//						"value": "terraform-vm-acctest-update",
			//					}),
			//				),
			//			},
			//			// Step 4 - Reset VM
			//			{
			//				Config: vmDependencies + ` `,
			//				Check:  resource.ComposeAggregateTestCheckFunc(),
			//			},
			//			// Step 5 - Create VM with attributes to update
			//			{
			//				Config: vmUpdateDependencies + `
			//resource "numspot_vm" "numspot-vm-acctest" {
			//  subnet_id          = numspot_subnet.terraform-dep-vm-subnet.id
			//  keypair_name       = numspot_keypair.terraform-dep-vm-keypair.name
			//  security_group_ids = [numspot_security_group.terraform-dep-vm-sg.id]
			//
			//  image_id                    = "ami-00669acb"
			//  type                        = "ns-eco6-2c2r"
			//  user_data                   = "dXNlci1kYXRhLWVuY29kZWQ="
			//  initiated_shutdown_behavior = "terminate"
			//
			//  tags = [{
			//    key   = "name"
			//    value = "terraform-vm-acctest"
			//  }]
			//}`,
			//				Check: resource.ComposeAggregateTestCheckFunc(
			//					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "subnet_id", "numspot_subnet.terraform-dep-vm-subnet", "id"),
			//					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "keypair_name", "numspot_keypair.terraform-dep-vm-keypair", "name"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "security_group_ids.#", "1"),
			//					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "security_group_ids.0", "numspot_security_group.terraform-dep-vm-sg", "id"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "image_id", "ami-00669acb"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "type", "ns-eco6-2c2r"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "user_data", "dXNlci1kYXRhLWVuY29kZWQ="),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "initiated_shutdown_behavior", "terminate"),
			//					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "tags.#", "1"),
			//					resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.numspot-vm-acctest", "tags.*", map[string]string{
			//						"key":   "name",
			//						"value": "terraform-vm-acctest",
			//					}),
			//				),
			//			},
			// Step 6 - Replace VM attributes
			{
				Config: vmUpdateDependencies + `
resource "numspot_vm" "numspot-vm-acctest" {
  subnet_id          = numspot_subnet.terraform-dep-vm-subnet.id
  keypair_name       = numspot_keypair.terraform-dep-vm-keypair.name
  security_group_ids = [numspot_security_group.terraform-dep-vm-sg.id]

  image_id                    = "ami-00669acb"
  type                        = "ns-eco6-2c2r"
  user_data                   = "dXNlci1kYXRhLWVuY29kZWQ="
  initiated_shutdown_behavior = "terminate"

  client_token  = "client-token"
  private_ips   = ["10.101.10.1"]
  placement = {
    tenancy                = "default"
    availability_zone_name = "cloudgouv-eu-west-1a"
  }

  tags = [{
    key   = "name"
    value = "terraform-vm-acctest-replace"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "subnet_id", "numspot_subnet.terraform-dep-vm-subnet", "id"),
					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "keypair_name", "numspot_keypair.terraform-dep-vm-keypair", "name"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest", "security_group_ids.0", "numspot_security_group.terraform-dep-vm-sg", "id"),

					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "image_id", "ami-00669acb"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "type", "ns-eco6-2c2r"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "user_data", "dXNlci1kYXRhLWVuY29kZWQ="),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "initiated_shutdown_behavior", "terminate"),

					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "client_token", "client-token"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "private_ips.#", "1"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "private_ips.0", "10.101.10.1"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "placement.tenancy", "default"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "placement.availability_zone_name", "cloudgouv-eu-west-1a"),

					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.numspot-vm-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-vm-acctest-replace",
					}),
				),
			},
			// Step 7 - Recreate VM with attributes to replace
			{
				Config: vmUpdateDependencies + `
resource "numspot_vm" "numspot-vm-acctest-recreate" {
  subnet_id          = numspot_subnet.terraform-dep-vm-subnet.id
  keypair_name       = numspot_keypair.terraform-dep-vm-keypair.name
  security_group_ids = [numspot_security_group.terraform-dep-vm-sg.id]

  image_id                    = "ami-00669acb"
  type                        = "ns-eco6-2c2r"
  user_data                   = "dXNlci1kYXRhLWVuY29kZWQ="
  initiated_shutdown_behavior = "stop"

  client_token  = "client-token"
  private_ips   = ["10.101.10.1"]
  placement = {
    tenancy                = "default"
    availability_zone_name = "cloudgouv-eu-west-1a"
  }

  tags = [{
    key   = "name"
    value = "terraform-vm-acctest-recreate"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest-recreate", "subnet_id", "numspot_subnet.terraform-dep-vm-subnet", "id"),
					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest-recreate", "keypair_name", "numspot_keypair.terraform-dep-vm-keypair", "name"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_vm.numspot-vm-acctest-recreate", "security_group_ids.0", "numspot_security_group.terraform-dep-vm-sg", "id"),

					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "image_id", "ami-00669acb"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "type", "ns-eco6-2c2r"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "user_data", "dXNlci1kYXRhLWVuY29kZWQ="),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "initiated_shutdown_behavior", "stop"),

					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "client_token", "client-token"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "private_ips.#", "1"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "private_ips.0", "10.101.10.1"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "placement.tenancy", "default"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "placement.availability_zone_name", "cloudgouv-eu-west-1a"),

					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-acctest-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_vm.numspot-vm-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-vm-acctest-recreate",
					}),
				),
			},
		},
	})
}
