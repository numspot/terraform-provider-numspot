resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2a"
}

data "numspot_volumes" "datasource_volume" {
  ids = [numspot_volume.volume.id]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_volumes.datasource-volume.items.0.id"
  }
}