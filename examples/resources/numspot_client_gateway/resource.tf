resource "numspot_client_gateway" "example" {
  connection_type = "ipsec.1"
  bgp_asn         = "192.0.2.0"
  public_ip       = 65000

  tags = [
    {
      key   = "Name"
      value = "My-Client-Gateway"
    }
  ]
}

