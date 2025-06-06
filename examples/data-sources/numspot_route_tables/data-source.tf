resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_internet_gateway" "internet-gateway" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_route_table" "route-table" {
  vpc_id    = numspot_vpc.vpc.id
  subnet_id = numspot_subnet.subnet.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.internet-gateway.id
    }
  ]
}

data "numspot_route_tables" "datasource-route-table" {
  ids = [numspot_route_table.route-table.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_route_tables.datasource-route-table.items.0.id"
  }
}