//go:build acc

package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccImageResource_FromImage(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	randint := rand.Intn(9999-1000) + 1000
	imageName := fmt.Sprintf("terraform-generated-volume-%d", randint)
	sourceImageId := "ami-026ce760"
	region := "cloudgouv-eu-west-1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testImageConfig_Create_FromImage(imageName, sourceImageId, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", imageName),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", sourceImageId),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", region),
					resource.TestCheckResourceAttr("numspot_image.test", "state", "available"),
				),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_image_id", "source_region_name"},
			},
		},
	})
}

func testImageConfig_Create_FromImage(name, sourceImageId, region string) string {
	return fmt.Sprintf(`
resource "numspot_image" "test" {
  name               = %[1]q
  source_image_id    = %[2]q
  source_region_name = %[3]q
}`, name, sourceImageId, region)
}

func TestAccImageResource_FromVm(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	randint := rand.Intn(9999-1000) + 1000
	imageName := fmt.Sprintf("terraform-generated-volume-%d", randint)
	sourceImageId := "ami-026ce760"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testImageConfig_Create_FromVm(sourceImageId, imageName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", imageName),
					resource.TestCheckResourceAttr("numspot_image.test", "state", "available"),
				),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vm_id"},
			},
		},
	})
}

func testImageConfig_Create_FromVm(imageId, name string) string {
	return fmt.Sprintf(`
resource "numspot_vm" "vm" {
  image_id = %[1]q
  vm_type  = "ns-cus6-2c4r"
}

resource "numspot_image" "test" {
  name  = %[2]q
  vm_id = numspot_vm.vm.id
}`, imageId, name)
}

func TestAccImageResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	randint := rand.Intn(9999-1000) + 1000
	imageName := fmt.Sprintf("terraform-generated-volume-%d", randint)
	sourceImageId := "ami-026ce760"
	region := "cloudgouv-eu-west-1"

	tagKey := "name"
	tagValue := "Terraform-Test-Image"
	tagValueUpdated := tagValue + "-Updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Create testing
			{
				Config: testImageConfig_Create_Tags(imageName, sourceImageId, region, tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", imageName),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", sourceImageId),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", region),
					resource.TestCheckResourceAttr("numspot_image.test", "state", "available"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
				),
				ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:            "numspot_image.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_image_id", "source_region_name"},
			},
			// Update testing
			{
				Config: testImageConfig_Create_Tags(imageName, sourceImageId, region, tagKey, tagValueUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", imageName),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", sourceImageId),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", region),
					resource.TestCheckResourceAttr("numspot_image.test", "state", "available"),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.0.value", tagValueUpdated),
					resource.TestCheckResourceAttr("numspot_image.test", "tags.#", "1"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testImageConfig_Create_Tags(name, sourceImageId, region, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "numspot_image" "test" {
  name               = %[1]q
  source_image_id    = %[2]q
  source_region_name = %[3]q

  tags = [
    {
      key   = %[4]q
      value = %[5]q
    }
  ]
}`, name, sourceImageId, region, tagKey, tagValue)
}
