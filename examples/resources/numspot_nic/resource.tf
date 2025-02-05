resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_nic" "nic" {
  subnet_id = numspot_subnet.subnet.id
  tags = [
    {
      key   = "name"
      value = "My Nic"
    }
  ]
}