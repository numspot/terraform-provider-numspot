package security_group_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestSecurityGroupResourceCreate(t *testing.T) {
	name := fmt.Sprintf("%s%d", "TEST_TERRAFORM_SECURITY_GROUP_NAME_", rand.Uint64())
	description := fmt.Sprintf("%s%d", "TEST_TERRAFORM_SECURITY_GROUP_DESCRIPTION_", rand.Uint64())

	updatedName := fmt.Sprintf("%s%s", name, "_UPDATED")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupConfigCreate(name, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.main", "name", name),
					resource.TestCheckResourceAttr("numspot_security_group.main", "description", description),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_security_group.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"virtual_private_cloud_id"},
			},
			// Update testing
			{
				Config: testSecurityGroupConfigCreate(updatedName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_security_group.main", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_security_group.main", "description", description),
				),
			},
		},
	})
}

func testSecurityGroupConfigCreate(name, description string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "vpc" {
	ip_range = "10.0.0.0/16"
}

resource "numspot_security_group" "main" {
	virtual_private_cloud_id 	= numspot_vpc.vpc.id
	name 						= %[1]q
	description					= %[2]q
}
`, name, description)
}
