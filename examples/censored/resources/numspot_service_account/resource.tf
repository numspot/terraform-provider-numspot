resource "numspot_service_account" "test" {
  space_id = "xxx-yyy-zzz"
  name     = "terraform-service-account"

  global_permissions = [
    "aaa-bbb-ccc",
  ]

  roles = [
    "aaa-xxx-zzz",
  ]
}