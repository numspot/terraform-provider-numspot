resource "numspot_service_account" "test" {
  space_id = "bba8c1df-609f-4775-9638-952d488502e6"
  name     = "My Service Account"
}

resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/24"
}

resource "numspot_acls" "acls_network" {
  space_id           = "bba8c1df-609f-4775-9638-952d488502e6"
  service_account_id = numspot_service_account.test.service_account_id
  service            = "network"
  resource           = "vpc"
  acls = [
    {
      resource_id   = numspot_vpc.vpc.id
      permission_id = "c537ba9e-19b6-4937-b7d4-7bf362d53bc6"
    }
  ]

}