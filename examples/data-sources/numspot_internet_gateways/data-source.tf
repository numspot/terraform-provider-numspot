resource "numspot_vpc" "vpc" {
  ip_range = "10.101.1.0/24"
  tags = [{
    key   = "name"
    value = "vpc"
  }]
}

resource "numspot_internet_gateway" "internet-gateway" {
  vpc_id = numspot_vpc.vpc.id
}

data "numspot_internet_gateways" "datasource-internet-gateway" {
  ids = [numspot_internet_gateway.internet-gateway.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_internet_gateways.datasource-internet-gateway.items.0.id"
  }
}