///go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccACLs(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	spaceID := "bba8c1df-609f-4775-9638-952d488502e6"

	service := "network"
	aclResource := "vpc"

	serviceVolume := "storageblock"
	resourceVolume := "volume"

	acls := `[
		{
		resource_id   = numspot_vpc.vpc.id
		permission_id = "c537ba9e-19b6-4937-b7d4-7bf362d53bc6"
		}
  	]`

	acls2 := `[
		{
		  resource_id   = numspot_vpc.vpc2.id
		  permission_id = "c537ba9e-19b6-4937-b7d4-7bf362d53bc6"
		}
	]`

	acls3 := `[
		{
		  resource_id   = numspot_volume.volume2.id
		  permission_id = "e572d5da-c6f8-4fdb-af5b-cab6ebaa628e"
		},
		{
		  resource_id   = numspot_volume.volume2.id
		  permission_id = "983b3c03-6b00-4862-9b16-7032df5a89ab"
		},
		{
		  resource_id   = numspot_volume.volume2.id
		  permission_id = "c3f5a8e2-248f-45dc-938b-be00aef9dc12"
		}
	]`

	acls4 := `[
		{
		  resource_id   = numspot_volume.volume.id
		  permission_id = "e572d5da-c6f8-4fdb-af5b-cab6ebaa628e"
		},
		{
		  resource_id   = numspot_volume.volume.id
		  permission_id = "983b3c03-6b00-4862-9b16-7032df5a89ab"
		},
		{
		  resource_id   = numspot_volume.volume.id
		  permission_id = "da7dd53a-f026-4534-a21d-51df7247f8dc"
		},
	]`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create ACLS
				Config: testACLsConfig(spaceID, service, aclResource, acls),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", service),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", aclResource),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "acls.0.resource_id", "numspot_vpc.vpc", "id"),
				),
			},
			{ // 2 - Update an ACLS
				Config: testACLsConfig(spaceID, service, aclResource, acls2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", service),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", aclResource),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "acls.0.resource_id", "numspot_vpc.vpc2", "id"),
				),
			},
			{ // 3 - Empty ACLS and delete dependencies (we need acls to be in "requiresReplace" mode to work)
				Config: testACLsConfig_NoSubResource(spaceID, service, aclResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", service),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", aclResource),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "0"),
				),
			},
			{ // 4 - Update acls's service/resource
				Config: testACLsConfigVolume(spaceID, serviceVolume, resourceVolume, acls3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", serviceVolume),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", resourceVolume),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_acls.acls_network", "acls.*", map[string]string{"permission_id": "e572d5da-c6f8-4fdb-af5b-cab6ebaa628e"}), // Note : we can't easily test value of permission_id and resource_id together (we need a mix of TestCheckTypeSetElemAttrPair and TestCheckTypeSetElemNestedAttrs function)
					resource.TestCheckTypeSetElemNestedAttrs("numspot_acls.acls_network", "acls.*", map[string]string{"permission_id": "983b3c03-6b00-4862-9b16-7032df5a89ab"}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_acls.acls_network", "acls.*", map[string]string{"permission_id": "c3f5a8e2-248f-45dc-938b-be00aef9dc12"}),
				),
			},
			{ // 5 - Update multiple acls
				Config: testACLsConfigVolume(spaceID, serviceVolume, resourceVolume, acls4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", serviceVolume),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", resourceVolume),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_acls.acls_network", "acls.*", map[string]string{"permission_id": "e572d5da-c6f8-4fdb-af5b-cab6ebaa628e"}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_acls.acls_network", "acls.*", map[string]string{"permission_id": "983b3c03-6b00-4862-9b16-7032df5a89ab"}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_acls.acls_network", "acls.*", map[string]string{"permission_id": "da7dd53a-f026-4534-a21d-51df7247f8dc"}),
				),
			},
		},
	})
}

func testACLsConfig(spaceID, service, resource, acls string) string {
	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = "My Service Account"
}

resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/24"
}

resource "numspot_vpc" "vpc2" {
  ip_range = "10.101.0.0/24"
}

resource "numspot_acls" "acls_network" {
  space_id           = %[1]q
  service_account_id = numspot_service_account.test.service_account_id
  service            = %[3]q
  resource           = %[4]q
  acls               = %[2]s

}`, spaceID, acls, service, resource)
}

func testACLsConfig_NoSubResource(spaceID, service, resource string) string {
	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = "My Service Account"
}


resource "numspot_acls" "acls_network" {
  space_id           = %[1]q
  service_account_id = numspot_service_account.test.service_account_id
  service            = %[2]q
  resource           = %[3]q
  acls               = []

}`, spaceID, service, resource)
}

func testACLsConfigVolume(spaceID, service, resource, acls string) string {
	return fmt.Sprintf(`
resource "numspot_service_account" "test" {
  space_id = %[1]q
  name     = "My Service Account"
}

resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = "11"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_volume" "volume2" {
  type                   = "standard"
  size                   = "22"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_acls" "acls_network" {
  space_id           = %[1]q
  service_account_id = numspot_service_account.test.service_account_id
  service            = %[3]q
  resource           = %[4]q
  acls               = %[2]s

}`, spaceID, acls, service, resource)
}
