resource "numspot_client_gateway" "client-gateway" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

data "numspot_client_gateways" "datasource-client-gateway" {
  depends_on = [numspot_client_gateway.client-gateway]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_client_gateways.datasource-client-gateway.items.0.id"
  }
}