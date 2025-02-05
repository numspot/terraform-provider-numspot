resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

data "numspot_volumes" "datasource_test" {
  ids = [numspot_volume.test.id]
}
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_volumes.testdata.items.0.id"
  }
}