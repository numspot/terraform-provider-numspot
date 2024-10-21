package test

//func TestAccServiceAccountDatasource(t *testing.T) {
//	// Due to a breaking change on IAM side this datasource becomes useless
//	// We skip this test in CI pipeline waiting for a re-design for GET /serviceAccounts endpoint from IAM side
//	t.Skip()
//	acct := acctest.NewAccTest(t, false, "")
//	defer func() {
//		err := acct.Cleanup()
//		require.NoError(t, err)
//	}()
//	pr := acct.TestProvider
//
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: pr,
//		Steps: []resource.TestStep{
//			{
//				Config: `
//resource "numspot_service_account" "test" {
//  space_id = "67d97ad4-3005-48dc-a392-60a97ab5097c"
//  name     = "terraform-service-account-test-datasource"
//  service_account_id =
//}
//
//data "numspot_service_accounts" "testdata" {
//  space_id            = numspot_service_account.test.space_id
//  service_account_ids = [numspot_service_account.test.service_account_id]
//}`,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("data.numspot_service_accounts.testdata", "items.#", "1"),
//					acctest.TestCheckTypeSetElemNestedAttrsWithPair("data.numspot_service_accounts.testdata", "items.*", map[string]string{
//						"id":   acctest.PAIR_PREFIX + "numspot_service_account.test.service_account_id",
//						"name": "terraform-service-account-test-datasource",
//					}),
//				),
//			},
//		},
//	})
//}
