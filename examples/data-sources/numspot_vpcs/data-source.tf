resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}
data "numspot_vpcs" "testdata" {
  ids = [numspot_vpc.test.id]
}

