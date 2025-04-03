resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2a"
}

resource "numspot_snapshot" "snapshot" {
  volume_id   = numspot_volume.volume.id
  description = "A beautiful snapshot"
  tags = [
    {
      key   = "name"
      value = "My Snapshot"
    }
  ]
}
