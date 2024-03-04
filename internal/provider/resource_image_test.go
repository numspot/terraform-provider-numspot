package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccImageResource_FromImage(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	randint := rand.Intn(9999-1000) + 1000
	imageName := fmt.Sprintf("terraform-generated-volume-%d", randint)
	sourceImageId := "ami-00b0c39a"
	region := "eu-west-2"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testImageConfig_Create_FromImage(imageName, sourceImageId, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_image.test", "name", imageName),
					resource.TestCheckResourceAttr("numspot_image.test", "source_image_id", sourceImageId),
					resource.TestCheckResourceAttr("numspot_image.test", "source_region_name", region),
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
	name 				= %[1]q
	source_image_id 	= %[2]q
	source_region_name	= %[3]q
}`, name, sourceImageId, region)
}
