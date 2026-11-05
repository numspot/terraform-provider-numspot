resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

data "numspot_subnets" "datasource-subnet" {
  vpc_ids    = [numspot_vpc.vpc.id]
  depends_on = [numspot_subnet.subnet]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_subnet.datasource-subnet.items.0.id"
  }
}