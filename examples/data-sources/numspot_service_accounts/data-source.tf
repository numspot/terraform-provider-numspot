resource "numspot_service_account" "test" {
  space_id = "00dc1f59-f473-4b73-82b6-67cab1b39d2d"
  name     = "My svc account"
}

data "numspot_service_accounts" "testdata" {
  space_id             = numspot_service_account.test.space_id
  service_account_name = numspot_service_account.test.name
}