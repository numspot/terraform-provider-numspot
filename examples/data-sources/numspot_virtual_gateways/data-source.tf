resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "virtual-gateway" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.vpc.id
}

data "numspot_virtual_gateways" "datasource-virtual-gateway" {
  depends_on = [numspot_virtual_gateway.virtual-gateway]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_virtual_gateways.datasource-virtual-gateway.items.0.id"
  }
}