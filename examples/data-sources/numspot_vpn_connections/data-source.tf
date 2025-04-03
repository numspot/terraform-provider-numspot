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

data "numspot_vpn_connections" "datasource-vpn-connection" {
  depends_on = [numspot_vpn_connection.vpn-connection]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_vpn_connections.datasource-vpn-connection.items.0.id"
  }
}