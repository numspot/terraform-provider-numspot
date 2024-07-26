resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  static_routes_only = false
}

data "numspot_vpn_connections" "testdata" {
  ids        = [numspot_vpn_connection.test.id]
  depends_on = [numspot_vpn_connection.test]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_vpn_connections.testdata.items.0.id"
  }
}