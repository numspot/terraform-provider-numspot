resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.test.id
}

data "numspot_internet_gateways" "testdata" {
  ids        = [numspot_internet_gateway.test.id]
  depends_on = [numspot_internet_gateway.test]
}