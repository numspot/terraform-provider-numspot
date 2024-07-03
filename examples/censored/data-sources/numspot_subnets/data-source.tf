data "numspot_subnets" "testdata" {
  vpc_ids    = [numspot_vpc.main.id]
  depends_on = [numspot_vpc.main, numspot_subnet.test]
}
resource "numspot_vpc" "main" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.main.id
  ip_range = "10.101.1.0/24"
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_subnet.testdata.items.0.id"
  }
}