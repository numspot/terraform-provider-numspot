resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id = numspot_volume.test.id
}

data "numspot_snapshots" "testdata" {
  ids = [numspot_snapshot.test.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_snapshots.testdata.items.0.id"
  }
}