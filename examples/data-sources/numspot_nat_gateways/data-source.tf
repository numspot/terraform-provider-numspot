resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "internet-gateway" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_subnet" "subnet" {
  vpc_id                  = numspot_vpc.vpc.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "public-ip" {}

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

resource "numspot_nat_gateway" "nat-gateway" {
  subnet_id    = numspot_subnet.subnet.id
  public_ip_id = numspot_public_ip.public-ip.id
  depends_on   = [numspot_route_table.route-table]
}

data "numspot_nat_gateways" "datasource-nat-gateway" {
  subnet_ids = [numspot_subnet.subnet.id]
  depends_on = [numspot_nat_gateway.nat-gateway]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_nat_gateways.datasource-nat-gateway.items.0.id"
  }
}