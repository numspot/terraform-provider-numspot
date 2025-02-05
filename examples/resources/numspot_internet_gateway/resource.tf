resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "internet_gateway" {
  vpc_id = numspot_vpc.vpc.id
  tags = [
    {
      key   = "name"
      value = "My Internet Gateway"
    }
  ]
}