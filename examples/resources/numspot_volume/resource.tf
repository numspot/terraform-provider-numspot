resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
}

# Tags
resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "foo"
      value = "bar"
    }
  ]
}