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