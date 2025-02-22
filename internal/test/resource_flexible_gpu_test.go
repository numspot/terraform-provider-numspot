package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccFlexibleGpuResource(t *testing.T) {
	// Due to Insufficient Capacity issues we can't run this test in a deterministic way
	// We skip this test in the CI until this problem is resolved
	t.Skip()
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create testing
				Config: `
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.net.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1a"
  vm_id                  = numspot_vm.test.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", "nvidia-a100-80"),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", "v6"),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_flexible_gpu.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing With Replace of VM resource and without Replacing the GPU
			// This test is useful to check wether or not the deletion of the VM and then the update of the GPU works properly
			{
				Config: `
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.net.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1b"
}

resource "numspot_vm" "test" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1a"
  vm_id                  = numspot_vm.test.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", "nvidia-a100-80"),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", "v6"),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", "cloudgouv-eu-west-1b"),
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
			// 4 - Update testing With Replace of VM resource and without Replacing the GPU
			// This test is useful to check wether or not the deletion of the VM and then the update of the GPU works properly
			{
				Config: `
resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id                 = numspot_vpc.net.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test_new" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.test.id
}

resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1a"
  vm_id                  = numspot_vm.test_new.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", "nvidia-a100-80"),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", "v6"),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},

			// <== If resource has optional dependencies ==>
			// 5 - Update testing With Delete of VM resource
			// This test is useful to check wether or not the deletion of the VM and then the update/replace of the GPU works properly (empty dependency)
			{
				Config: `
resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1a"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", "nvidia-a100-80"),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", "v6"),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", "cloudgouv-eu-west-1a"),
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
