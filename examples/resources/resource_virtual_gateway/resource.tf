resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"

  tags = [
    {
      key   = "name"
      value = "terraform-virtual-gateway"
    }
  ]
}