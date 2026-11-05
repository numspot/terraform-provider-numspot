//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

// TestAccVmResourceSgOnlyHotUpdate covers CLOS-3261 (and completes SUPN2-643 /
// SUPN2-718 coverage): associating/dissociating a security group on a running
// VM must update the VM in place without any stop/start API call, without
// replacing the VM, and must preserve an explicit block_device_mappings
// bsu.volume_size (SUPN2-718 client scenario).
//
// The recorded cassette contains no PUT .../stop or PUT .../start interaction,
// so replaying it fails if the provider ever reintroduces the stop/start
// cycle for a security-group-only change.
func TestAccVmResourceSgOnlyHotUpdate(t *testing.T) {
	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	dependencies := `
resource "numspot_vpc" "terraform-dep-vm-hot-sg-vpc" {
  ip_range = "10.104.0.0/16"
  tags = [{
    key   = "name"
    value = "terraform-dep-vm-hot-sg-vpc"
  }]
}

resource "numspot_subnet" "terraform-dep-vm-hot-sg-subnet" {
  vpc_id                 = numspot_vpc.terraform-dep-vm-hot-sg-vpc.id
  ip_range               = "10.104.1.0/24"
  availability_zone_name = "eu-west-2a"
  tags = [{
    key   = "name"
    value = "terraform-dep-vm-hot-sg-subnet"
  }]
}

resource "numspot_security_group" "terraform-dep-vm-hot-sg-base" {
  vpc_id      = numspot_vpc.terraform-dep-vm-hot-sg-vpc.id
  name        = "terraform-dep-vm-hot-sg-base"
  description = "terraform-dep-vm-hot-sg-base-description"
}

resource "numspot_security_group" "terraform-dep-vm-hot-sg-extra" {
  vpc_id      = numspot_vpc.terraform-dep-vm-hot-sg-vpc.id
  name        = "terraform-dep-vm-hot-sg-extra"
  description = "terraform-dep-vm-hot-sg-extra-description"
}
`

	vmConfig := func(sgRefs string) string {
		return `
` + dependencies + `
resource "numspot_vm" "numspot-vm-hot-sg-acctest" {
  subnet_id          = numspot_subnet.terraform-dep-vm-hot-sg-subnet.id
  image_id           = "ami-88dbc914"
  type               = "ns-cus6-2c4r"

  security_group_ids = [` + sgRefs + `]

  block_device_mappings = [{
    device_name = "/dev/sda1"
    bsu = {
      volume_size = 30
      volume_type = "gp2"
    }
  }]

  tags = [{
    key   = "name"
    value = "terraform-vm-hot-sg-acctest"
  }]
}`
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Step 1 - Create running VM with an explicit block device mapping
			// (custom root volume size) and the base security group.
			{
				Config: vmConfig(`numspot_security_group.terraform-dep-vm-hot-sg-base.id`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-hot-sg-acctest", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-hot-sg-acctest", "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-hot-sg-acctest", "block_device_mappings.0.bsu.volume_size", "30"),
				),
			},
			// Step 2 - Associate the extra security group: in-place update,
			// no stop/start, volume_size preserved (SUPN2-718).
			{
				Config: vmConfig(`numspot_security_group.terraform-dep-vm-hot-sg-base.id, numspot_security_group.terraform-dep-vm-hot-sg-extra.id`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-hot-sg-acctest", "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-hot-sg-acctest", "block_device_mappings.0.bsu.volume_size", "30"),
				),
			},
			// Step 3 - Dissociate the extra security group: same expectations.
			{
				Config: vmConfig(`numspot_security_group.terraform-dep-vm-hot-sg-base.id`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-hot-sg-acctest", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("numspot_vm.numspot-vm-hot-sg-acctest", "block_device_mappings.0.bsu.volume_size", "30"),
				),
			},
		},
	})
}
