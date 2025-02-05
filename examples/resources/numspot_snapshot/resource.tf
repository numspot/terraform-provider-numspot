resource "numspot_volume" "test" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_snapshot" "test" {
  volume_id   = numspot_volume.test.id
  description = "A beautiful snapshot"
  tags = [
    {
      key   = "name"
      value = "My Snapshot"
    }
  ]
}
