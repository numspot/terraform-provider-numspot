resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

data "numspot_client_gateways" "testdata" {
  ids        = [numspot_client_gateway.test.id]
  depends_on = [numspot_client_gateway.test]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_client_gateways.testdata.items.0.id"
  }
}