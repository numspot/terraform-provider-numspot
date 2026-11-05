resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

data "numspot_vpcs" "datasource-vpc" {
  ids = [numspot_vpc.vpc.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_vpcs.datasource-vpc.items.0.id"
  }
}