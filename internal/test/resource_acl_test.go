package test

//
//import (
//	"fmt"
//	"slices"
//	"strconv"
//	"strings"
//	"testing"
//
//	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//	"github.com/stretchr/testify/require"
//
//	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
//)
//
//// This struct will store the input data that will be used in your tests (all fields as string)
//type StepDataACLs struct {
//	spaceId,
//	service,
//	resource string
//	actions []string
//}
//
//// Generate checks to validate that resource 'numspot_acls.test' has input data values
//func getFieldMatchChecksACLs(data StepDataACLs) []resource.TestCheckFunc {
//	return []resource.TestCheckFunc{
//		resource.TestCheckResourceAttr("numspot_acls.test", "space_id", data.spaceId),
//		resource.TestCheckResourceAttr("numspot_acls.test", "service", data.service),
//		resource.TestCheckResourceAttr("numspot_acls.test", "resource", data.resource),
//	}
//}
//
//// Generate checks to validate that resource 'numspot_acls.test' is properly linked to given subresources
//// If resource has no dependencies, return empty array
//func getDependencyChecksACLs(dependenciesSuffix string, data StepDataACLs) []resource.TestCheckFunc {
//	checks := []resource.TestCheckFunc{
//		resource.TestCheckResourceAttrPair("numspot_acls.test", "service_account_id", "numspot_service_account.test"+dependenciesSuffix, "service_account_id"),
//		resource.TestCheckResourceAttr("numspot_acls.test", "acls.#", strconv.Itoa(len(data.actions))),
//	}
//
//	for _, action := range data.actions {
//		checks = append(checks, resource.TestCheckTypeSetElemNestedAttrs("numspot_acls.test", "acls.*", map[string]string{
//			"permission_id": fmt.Sprintf(provider.PAIR_PREFIX+"data.numspot_permissions.%s.items.0.id", generatePermissionName(action, data.service, data.resource)),
//			"resource_id":   fmt.Sprintf(provider.PAIR_PREFIX+"numspot_%s.test%s.id", data.resource, dependenciesSuffix),
//		}))
//	}
//
//	return checks
//}
//
//func TestAccACLsResource(t *testing.T) {
//	pr := provider.TestAccProtoV6ProviderFactories
//
//	var resourceId string
//
//	spaceID := "bba8c1df-609f-4775-9638-952d488502e6"
//
//	////////////// Define input data that will be used in the test sequence //////////////
//	// resource fields that can be updated in-place
//	// None
//
//	// resource fields that cannot be updated in-place (requires replace)
//	service := "network"
//	serviceUpdated := "storageblock"
//
//	resourceName := "vpc"
//	resourceNameUpdated := "volume"
//
//	actions := []string{"getIAMPolicy"}
//	actionsUpdated1 := []string{"update", "unlink", "getIAMPolicy"}
//	actionsUpdated2 := []string{"update", "unlink", "create"}
//
//	/////////////////////////////////////////////////////////////////////////////////////
//
//	////////////// Define plan values and generate associated attribute checks  //////////////
//	// The base plan (used in first create and to reset resource state before some tests)
//	basePlanValues := StepDataACLs{
//		spaceId:  spaceID,
//		service:  service,
//		resource: resourceName,
//		actions:  actions,
//	}
//	createChecks := append(
//		getFieldMatchChecksACLs(basePlanValues),
//
//		resource.TestCheckResourceAttrWith("numspot_acls.test", "id", func(v string) error {
//			if !assert.NotEmpty(t, v) {
//							return fmt.Errorf("Id field should not be empty")
//						}
//			// resourceId = v
//			return nil
//		}),
//	)
//
//	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
//	replacePlanValues1 := StepDataACLs{
//		spaceId:  spaceID,
//		service:  serviceUpdated,
//		resource: resourceNameUpdated,
//		actions:  actionsUpdated1,
//	}
//	replaceChecks1 := append(
//		getFieldMatchChecksACLs(replacePlanValues1),
//
//		resource.TestCheckResourceAttrWith("numspot_acls.test", "id", func(v string) error {
//			if !assert.NotEmpty(t, v) {
//							return fmt.Errorf("Id field should not be empty")
//						}
//			if !assert.NotEqual(t, resourceId, v) {
//							return fmt.Errorf("Id should have changed")
//						}
//			return nil
//		}),
//	)
//
//	replacePlanValues2 := StepDataACLs{
//		spaceId:  spaceID,
//		service:  serviceUpdated,
//		resource: resourceNameUpdated,
//		actions:  actionsUpdated2,
//	}
//	replaceChecks2 := append(
//		getFieldMatchChecksACLs(replacePlanValues2),
//
//		resource.TestCheckResourceAttrWith("numspot_acls.test", "id", func(v string) error {
//			if !assert.NotEmpty(t, v) {
//							return fmt.Errorf("Id field should not be empty")
//						}
//			if !assert.NotEqual(t, resourceId, v) {
//							return fmt.Errorf("Id should have changed")
//						}
//			return nil
//		}),
//	)
//	/////////////////////////////////////////////////////////////////////////////////////
//
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: pr,
//		Steps: []resource.TestStep{
//			{ // Create testing
//				Config: testACLsConfig(provider.BASE_SUFFIX, basePlanValues),
//				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
//					createChecks,
//					getDependencyChecksACLs(provider.BASE_SUFFIX, basePlanValues),
//				)...),
//			},
//			// ImportState testing
//			{
//				ResourceName:            "numspot_acls.test",
//				ImportState:             true,
//				ImportStateVerify:       true,
//				ImportStateVerifyIgnore: []string{"id"},
//			},
//			// Update testing With Replace (if needed)
//			{
//				Config: testACLsConfig(provider.BASE_SUFFIX, replacePlanValues1),
//				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
//					replaceChecks1,
//					getDependencyChecksACLs(provider.BASE_SUFFIX, replacePlanValues1),
//				)...),
//			},
//			// Update testing With Replace (if needed)
//			{
//				Config: testACLsConfig(provider.BASE_SUFFIX, replacePlanValues2),
//				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
//					replaceChecks2,
//					getDependencyChecksACLs(provider.BASE_SUFFIX, replacePlanValues2),
//				)...),
//			},
//
//			// <== If resource has required dependencies ==>
//			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
//				Config: testACLsConfig(provider.BASE_SUFFIX, basePlanValues),
//			},
//			// Update testing With Replace of dependency resource and with Replace of the resource (if needed)
//			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
//			{
//				Config: testACLsConfig(provider.NEW_SUFFIX, replacePlanValues1),
//				Check: resource.ComposeAggregateTestCheckFunc(slices.Concat(
//					replaceChecks1,
//					getDependencyChecksACLs(provider.NEW_SUFFIX, replacePlanValues1),
//				)...),
//			},
//
//			// <== If resource has optional dependencies ==>
//			{ // Reset the resource to initial state (resource tied to a subresource) in prevision of next test
//				Config: testACLsConfig(provider.BASE_SUFFIX, basePlanValues),
//			},
//			// Update testing With Deletion of dependency resource and with Replace of the resource (if needed)
//			// This test is useful to check wether or not the deletion of the dependencies and then the replace of the main resource works properly (empty dependency)
//			{
//				Config: testACLsConfig_DeletedDependencies(replacePlanValues1),
//				Check:  resource.ComposeAggregateTestCheckFunc(replaceChecks1...),
//			},
//		},
//	})
//}
//
//func testACLsConfig(subresourceSuffix string, data StepDataACLs) string {
//	var aclList, permissionDatasources string
//	for _, action := range data.actions {
//		permissionName := generatePermissionName(action, data.service, data.resource)
//
//		permissionDatasources += getPermissionHCLBLock(data.spaceId, action, data.service, data.resource, permissionName)
//
//		aclList += getACLBlock(fmt.Sprintf("numspot_%[1]s.test%[2]s.id", data.resource, subresourceSuffix), permissionName)
//		aclList += ","
//	}
//
//	aclList = strings.TrimSuffix(aclList, ",")
//
//	return fmt.Sprintf(`
//resource "numspot_service_account" "test%[1]s" {
//  space_id = %[2]q
//  name     = "My Service Account"
//}
//
//resource "numspot_vpc" "test%[1]s" {
//  ip_range = "10.101.0.0/24"
//}
//
//resource "numspot_volume" "test%[1]s" {
//  type                   = "standard"
//  size                   = "11"
//  availability_zone_name = "cloudgouv-eu-west-1a"
//}
//
//// list of permission datasources from list of action
//%[5]s
//
//resource "numspot_acls" "test" {
//  space_id           = %[2]q
//  service_account_id = numspot_service_account.test.service_account_id
//  service            = %[3]q
//  resource           = %[4]q
//  acls = [
//    %[6]s
//  ]
//
//}`, subresourceSuffix, data.spaceId, data.service, data.resource, permissionDatasources, aclList)
//}
//
//// <== If resource has optional dependencies ==>
//func testACLsConfig_DeletedDependencies(data StepDataACLs) string {
//	return fmt.Sprintf(`
//resource "numspot_service_account" "test" {
//  space_id = %[1]q
//  name     = "My Service Account"
//}
//
//resource "numspot_acls" "test" {
//  space_id           = %[1]q
//  service_account_id = numspot_service_account.test.service_account_id
//  service            = %[2]q
//  resource           = %[3]q
//}`, data.spaceId, data.service, data.resource)
//}
//
//func generatePermissionName(action, service, resource string) string {
//	return fmt.Sprintf("perm_%[1]s_%[2]s_%[3]s", action, service, resource)
//}
//
//func getPermissionHCLBLock(spaceId, action, service, resource, permissionName string) string {
//	return fmt.Sprintf(`
//data "numspot_permissions" "%[5]s" {
//  space_id = %[1]q
//  action   = %[2]q
//  service  = %[3]q
//  resource = %[4]q
//}`, spaceId, action, service, resource, permissionName)
//}
//
//func getACLBlock(resourceId, permissionName string) string {
//	return fmt.Sprintf(`{
//      resource_id   = %[1]s
//      permission_id = data.numspot_permissions.%[2]s.items.0.id
//    }
//`, resourceId, permissionName)
//}
