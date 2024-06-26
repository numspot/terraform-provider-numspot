///go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type (
	stepData struct {
		service    string
		resource   string
		actions    []string
		resourceID string
	}
)

func TestAccACLsResource(t *testing.T) {
	testData := []stepData{
		{
			service:    "network",
			resource:   "vpc",
			actions:    []string{"getIAMPolicy"},
			resourceID: "numspot_vpc.vpc.id",
		},
		{
			service:    "network",
			resource:   "vpc",
			actions:    []string{"getIAMPolicy"},
			resourceID: "numspot_vpc.vpc2.id",
		},
		{
			service:  "network",
			resource: "vpc",
		},
		{
			service:    "storageblock",
			resource:   "volume",
			actions:    []string{"update", "unlink", "getIAMPolicy"},
			resourceID: "numspot_volume.volume2.id",
		},
		{
			service:    "storageblock",
			resource:   "volume",
			actions:    []string{"update", "unlink", "create"},
			resourceID: "numspot_volume.volume.id",
		},
	}
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	// Required
	spaceID := "68134f98-205b-4de4-8b85-f6a786ef6481"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create ACLS
				Config: testACLsConfig(spaceID, testData[0]),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", testData[0].service),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", testData[0].resource),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "acls.0.resource_id", "numspot_vpc.vpc", "id"),
				),
			},
			{ // 2 - Update an ACLS
				Config: testACLsConfig(spaceID, testData[1]),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", testData[1].service),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", testData[1].resource),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "1"),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "acls.0.resource_id", "numspot_vpc.vpc2", "id"),
				),
			},
			{ // 3 - Empty ACLS and delete dependencies (we need acls to be in "requiresReplace" mode to work)
				Config: testACLsConfig_NoSubResource(spaceID, testData[2]),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", testData[2].service),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", testData[2].resource),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "0"),
				),
			},
			{ // 4 - Update acls's service/resource
				Config: testACLsConfigVolume(spaceID, testData[3]),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", testData[3].service),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", testData[3].resource),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "3"),
					resource.TestCheckTypeSetElemAttrPair("numspot_acls.acls_network", "acls.*.permission_id", "data.numspot_permissions.perm_1", "items.0.id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_acls.acls_network", "acls.*.permission_id", "data.numspot_permissions.perm_2", "items.0.id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_acls.acls_network", "acls.*.permission_id", "data.numspot_permissions.perm_3", "items.0.id"),
				),
			},
			{ // 5 - Update multiple acls
				Config: testACLsConfigVolume(spaceID, testData[4]),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "space_id", spaceID),
					resource.TestCheckResourceAttrPair("numspot_acls.acls_network", "service_account_id", "numspot_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "service", testData[4].service),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "resource", testData[4].resource),
					resource.TestCheckResourceAttr("numspot_acls.acls_network", "acls.#", "3"),
					resource.TestCheckTypeSetElemAttrPair("numspot_acls.acls_network", "acls.*.permission_id", "data.numspot_permissions.perm_1", "items.0.id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_acls.acls_network", "acls.*.permission_id", "data.numspot_permissions.perm_2", "items.0.id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_acls.acls_network", "acls.*.permission_id", "data.numspot_permissions.perm_3", "items.0.id"),
				),
			},
		},
	})
}

func testACLsConfig(spaceID string, testData stepData) string {
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

data "numspot_permissions" "perm_getIAMPolicy_vpc" {
  space_id = %[1]q
  action   = %[4]q
  service  = %[2]q
  resource = %[3]q
}

resource "numspot_acls" "acls_network" {
  space_id           = %[1]q
  service_account_id = numspot_service_account.test.service_account_id
  service            = %[2]q
  resource           = %[3]q
  acls = [
    {
      resource_id   = %[5]s
      permission_id = data.numspot_permissions.perm_getIAMPolicy_vpc.items.0.id
    }
  ]

}`, spaceID, testData.service, testData.resource, testData.actions[0], testData.resourceID)
}

func testACLsConfig_NoSubResource(spaceID string, testData stepData) string {
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

}`, spaceID, testData.service, testData.resource)
}

func testACLsConfigVolume(spaceID string, testData stepData) string {
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

data "numspot_permissions" "perm_1" {
  space_id = %[1]q
  action   = %[4]q
  service  = %[2]q
  resource = %[3]q
}

data "numspot_permissions" "perm_2" {
  space_id = %[1]q
  action   = %[5]q
  service  = %[2]q
  resource = %[3]q
}

data "numspot_permissions" "perm_3" {
  space_id = %[1]q
  action   = %[6]q
  service  = %[2]q
  resource = %[3]q
}

resource "numspot_acls" "acls_network" {
  space_id           = %[1]q
  service_account_id = numspot_service_account.test.service_account_id
  service            = %[2]q
  resource           = %[3]q
  acls = [
    {
      resource_id   = %[7]s
      permission_id = data.numspot_permissions.perm_1.items.0.id
    },
    {
      resource_id   = %[7]s
      permission_id = data.numspot_permissions.perm_2.items.0.id
    },
    {
      resource_id   = %[7]s
      permission_id = data.numspot_permissions.perm_3.items.0.id
    }
  ]

}`, spaceID,
		testData.service,
		testData.resource,
		testData.actions[0],
		testData.actions[1],
		testData.actions[2],
		testData.resourceID)
}
