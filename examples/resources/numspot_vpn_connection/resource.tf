resource "numspot_client_gateway" "client-gateway" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.0.0"
  bgp_asn         = 123456
}

resource "numspot_virtual_gateway" "virtual-gateway" {
  connection_type = "ipsec.1"
}

resource "numspot_vpn_connection" "vpn-connection" {
  client_gateway_id  = numspot_client_gateway.client-gateway.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.virtual-gateway.id
  static_routes_only = false
}
