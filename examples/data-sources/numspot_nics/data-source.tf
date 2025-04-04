resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}


resource "numspot_nic" "nic" {
  subnet_id = numspot_subnet.subnet.id
}

data "numspot_nics" "datasource-nic" {
  ids = [numspot_nic.nic.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_nics.datasource-nic.items.0.id"
  }
}