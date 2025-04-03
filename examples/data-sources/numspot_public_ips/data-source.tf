resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "internet-gateway" {
  vpc_id     = numspot_vpc.vpc.id
  depends_on = [numspot_nic.nic]
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "nic" {
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_public_ip" "public-ip" {
  nic_id     = numspot_nic.nic.id
  depends_on = [numspot_internet_gateway.internet-gateway]
}

data "numspot_public_ips" "datasource-public-ip" {
  ids = [numspot_public_ip.public-ip.id]
}