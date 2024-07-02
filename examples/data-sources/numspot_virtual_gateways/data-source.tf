resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
}

data "numspot_virtual_gateways" "testdata" {
  ids = [numspot_virtual_gateway.test.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_virtual_gateways.testdata.items.0.id"
  }
}