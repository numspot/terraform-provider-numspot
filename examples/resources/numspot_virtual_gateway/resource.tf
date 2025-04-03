resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "virtual-gateway" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.vpc.id
}

data "numspot_virtual_gateways" "datasource-virtual-gateways-acctest" {
  depends_on = [numspot_virtual_gateway.virtual-gateway]
}