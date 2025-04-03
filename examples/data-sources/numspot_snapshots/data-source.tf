resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2a"
}

resource "numspot_snapshot" "snapshot" {
  volume_id = numspot_volume.volume.id
}

data "numspot_snapshots" "datasource-snapshot" {
  ids = [numspot_snapshot.snapshot.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_snapshots.datasource-snapshots.items.0.id"
  }
}