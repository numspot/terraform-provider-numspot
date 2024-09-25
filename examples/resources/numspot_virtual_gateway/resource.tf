resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "vg" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.vpc.id

  tags = [
    {
      key   = "name"
      value = "My Virtual Gateway"
    }
  ]
}