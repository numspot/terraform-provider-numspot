//go:build acc

package test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

// TestAccPublicIpResource_VMUpdatePreservesPublicIP covers SUPN2-643:
// associating/dissociating a security group on a numspot_vm must not replace
// the numspot_public_ip resource (which used to rotate the public IP) and the
// public IP address itself must remain the same.
//
// Its cassette cannot be recorded yet: on the revamped main, real VM creation
// fails on a schema/type mismatch (serializeNic writes a flat string list for
// nics.private_ips where the schema expects nested PrivateIps objects). Until
// that regression is fixed and the cassette is committed, the test is skipped
// in replay mode so the rest of the suite is unaffected. To record it once VM
// creation works: make record-cassettes TESTARGS="-run TestAccPublicIpResource_VMUpdatePreservesPublicIP"
func TestAccPublicIpResource_VMUpdatePreservesPublicIP(t *testing.T) {
	if os.Getenv("VCR_MODE") == "replay" {
		cassette := filepath.Join("testdata", "TestAccPublicIpResource-VMUpdatePreservesPublicIP.cassette.yaml")
		if _, err := os.Stat(cassette); errors.Is(err, os.ErrNotExist) {
			t.Skipf("cassette not found (%s): skipping in replay mode until it is recorded", cassette)
		}
	}

	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var publicIPBefore string

	vmDependencies := `
resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.test.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "eu-west-2a"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

resource "numspot_security_group" "sg_base" {
  vpc_id      = numspot_vpc.test.id
  name        = "tf-acctest-sg-base"
  description = "base security group for SUPN2-643 regression test"
}

resource "numspot_security_group_rule" "sg_base_in" {
  security_group_id = numspot_security_group.sg_base.id
  flow              = "Inbound"
  from_port_range   = 22
  to_port_range     = 22
  ip_protocol       = "tcp"
  ip_range          = "0.0.0.0/0"
}

resource "numspot_security_group" "sg_test" {
  vpc_id      = numspot_vpc.test.id
  name        = "tf-acctest-sg-test"
  description = "toggled security group for SUPN2-643 regression test"
}

resource "numspot_security_group_rule" "sg_test_in" {
  security_group_id = numspot_security_group.sg_test.id
  flow              = "Inbound"
  from_port_range   = 80
  to_port_range     = 80
  ip_protocol       = "tcp"
  ip_range          = "0.0.0.0/0"
}
`

	vmBaseSG := vmDependencies + `
resource "numspot_vm" "test" {
  image_id           = "ami-43c6241c"
  type               = "ns-cus6-2c4r"
  subnet_id          = numspot_subnet.test.id
  security_group_ids = [numspot_security_group.sg_base.id]
}

resource "numspot_public_ip" "test" {
  vm_id      = numspot_vm.test.id
  depends_on = [numspot_internet_gateway.test]
}
`

	vmBaseAndExtraSG := vmDependencies + `
resource "numspot_vm" "test" {
  image_id           = "ami-43c6241c"
  type               = "ns-cus6-2c4r"
  subnet_id          = numspot_subnet.test.id
  security_group_ids = [numspot_security_group.sg_base.id, numspot_security_group.sg_test.id]
}

resource "numspot_public_ip" "test" {
  vm_id      = numspot_vm.test.id
  depends_on = [numspot_internet_gateway.test]
}
`

	capturePublicIP := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources["numspot_public_ip.test"]
			if !ok {
				return fmt.Errorf("resource numspot_public_ip.test not found in state")
			}
			publicIPBefore = rs.Primary.Attributes["public_ip"]
			return nil
		}
	}

	checkPublicIPUnchanged := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources["numspot_public_ip.test"]
			if !ok {
				return fmt.Errorf("resource numspot_public_ip.test not found in state")
			}
			current := rs.Primary.Attributes["public_ip"]
			if current != publicIPBefore {
				t.Errorf("public IP changed from %q to %q: the numspot_public_ip resource was replaced (SUPN2-643)", publicIPBefore, current)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - create VM with base SG + linked public IP
				Config: vmBaseSG,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("numspot_public_ip.test", "vm_id", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttrSet("numspot_public_ip.test", "public_ip"),
					capturePublicIP(),
				),
			},
			{ // 2 - associate the extra security group: public IP must not be replaced
				Config: vmBaseAndExtraSG,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("numspot_vm.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("numspot_public_ip.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPublicIPUnchanged(),
				),
			},
			{ // 3 - dissociate the extra security group: public IP must not be replaced
				Config: vmBaseSG,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("numspot_vm.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("numspot_public_ip.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPublicIPUnchanged(),
				),
			},
		},
	})
}
