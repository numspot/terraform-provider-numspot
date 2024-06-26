resource "numspot_service_account" "test" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "My Service Account"
}

resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/24"
}

data "numspot_permissions" "get_vpc_perm" {
  space_id = "68134f98-205b-4de4-8b85-f6a786ef6481"
  action   = "get"
  service  = "network"
  resource = "link"
}

resource "numspot_acls" "acls_network" {
  space_id           = "bba8c1df-609f-4775-9638-952d488502e6"
  service_account_id = numspot_service_account.test.service_account_id
  service            = "network"
  resource           = "vpc"
  acls = [
    {
      resource_id   = numspot_vpc.vpc.id
      permission_id = data.numspot_permissions.get_vpc_perm.items.0.id
    }
  ]

}