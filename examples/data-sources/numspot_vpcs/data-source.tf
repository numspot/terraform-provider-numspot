resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

data "numspot_vpcs" "testdata" {
  ids = [numspot_vpc.test.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_vpcs.testdata.items.0.id"
  }
}