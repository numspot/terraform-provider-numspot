//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFlexibleGpuResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	var flexibleGPUID string
	flexibleGpuModelName := "nvidia-a100-80"
	flexibleGpuGeneration := "v6"
	flexibleGpuAZ := "cloudgouv-eu-west-1a"
	deleteOnVMDeletion := "false"
	deleteOnVMDeletionUpdated := "true"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testFlexibleGpuConfig(flexibleGpuModelName, flexibleGpuGeneration, flexibleGpuAZ, deleteOnVMDeletion),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
						flexibleGPUID = v
						require.NotEmpty(t, v)
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", flexibleGpuModelName),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", flexibleGpuGeneration),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", flexibleGpuAZ),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "delete_on_vm_deletion", deleteOnVMDeletion),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_flexible_gpu.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testFlexibleGpuConfig(flexibleGpuModelName, flexibleGpuGeneration, flexibleGpuAZ, deleteOnVMDeletionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
						require.Equal(t, flexibleGPUID, v)
						return nil
					}),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "model_name", flexibleGpuModelName),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "generation", flexibleGpuGeneration),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "availability_zone_name", flexibleGpuAZ),
					resource.TestCheckResourceAttr("numspot_flexible_gpu.test", "delete_on_vm_deletion", deleteOnVMDeletionUpdated),
				),
			},
		},
	})
}

func testFlexibleGpuConfig(modelName, generation, az string, deleteOnVMDeletion string) string {
	return fmt.Sprintf(`
resource "numspot_flexible_gpu" "test" {
  model_name             = %[1]q
  generation             = %[2]q
  availability_zone_name = %[3]q
  delete_on_vm_deletion  = %[4]q
}`, modelName, generation, az, deleteOnVMDeletion)
}

func TestAccFlexibleGpuResourceLink(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	var flexibleGPUID string

	vmSuffix := "1"
	vmSuffixUpdated := "2"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testFlexibleGpuConfigLink(vmSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
						flexibleGPUID = v
						require.NotEmpty(t, v)
						return nil
					}),
					resource.TestCheckResourceAttrPair("numspot_flexible_gpu.test", "vm_id", fmt.Sprintf("numspot_vm.test%[1]s", vmSuffix), "id"),
				),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_flexible_gpu.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testFlexibleGpuConfigLink(vmSuffixUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_flexible_gpu.test", "id", func(v string) error {
						require.Equal(t, flexibleGPUID, v)
						return nil
					}),
					resource.TestCheckResourceAttrPair("numspot_flexible_gpu.test", "vm_id", fmt.Sprintf("numspot_vm.test%[1]s", vmSuffixUpdated), "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testFlexibleGpuConfigLink(vmSuffix string) string {
	return fmt.Sprintf(`
resource "numspot_vm" "test1" {
  image_id = "ami-026ce760"
  type     = "ns-mus6-2c16r"
}

resource "numspot_vm" "test2" {
  image_id = "ami-026ce760"
  type     = "ns-mus6-2c16r"
}

resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1a"
  vm_id                  = numspot_vm.test%[1]s.id
}`, vmSuffix)
}
