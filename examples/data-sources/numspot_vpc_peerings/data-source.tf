resource "numspot_vpc" "source" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "accepter" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter.id
  source_vpc_id   = numspot_vpc.source.id
}

data "numspot_vpc_peerings" "testdata" {
  ids = [numspot_vpc_peering.test.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_vpc_peerings.testdata.items.0.id"
  }
}