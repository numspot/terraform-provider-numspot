resource "numspot_net" "net" {
  ip_range = "10.0.0.0/16"
}

# Create a private subnet
resource "numspot_subnet" "private" {
  net_id   = numspot_net.net.id
  ip_range = "10.0.1.0/24"
}

# Create a public subnet (should have an internet_service)
resource "numspot_subnet" "public" {
  net_id                  = numspot_net.net.id
  ip_range                = "10.0.2.0/24"
  map_public_ip_on_launch = true
}