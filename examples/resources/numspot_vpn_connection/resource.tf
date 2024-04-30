resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  static_routes_only = true

  tags = [
    {
      key   = "Name"
      value = "Terraform VPN Connection"
    }
  ]
}