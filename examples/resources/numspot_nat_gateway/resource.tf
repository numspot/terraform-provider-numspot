resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "ig" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_subnet" "subnet" {
  vpc_id                  = numspot_vpc.vpc.id
  map_public_ip_on_launch = true
  ip_range                = "10.101.1.0/24"
}

resource "numspot_public_ip" "public_ip" {}

resource "numspot_route_table" "rt" {
  vpc_id    = numspot_vpc.vpc.id
  subnet_id = numspot_subnet.subnet.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.ig.id
    }
  ]
}

resource "numspot_nat_gateway" "ng" {
  subnet_id    = numspot_subnet.subnet.id
  public_ip_id = numspot_public_ip.public_ip.id
  tags = [
    {
      key   = "name"
      value = "My-Nat-Gateway"
    }
  ]

  depends_on = [numspot_route_table.rt]
}