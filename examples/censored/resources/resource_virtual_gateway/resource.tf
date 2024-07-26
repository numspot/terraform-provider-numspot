resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id

  tags = [
    {
      key   = "name"
      value = "terraform-virtual-gateway"
    }
  ]
}