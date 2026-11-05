resource "numspot_vpc" "vpc" {
  ip_range = "10.0.0.0/16"
}

# Create a private subnet
resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.0.1.0/24"
}

# Create a public subnet (should have an internet_service)
resource "numspot_subnet" "subnet" {
  vpc_id                  = numspot_vpc.vpc.id
  ip_range                = "10.0.2.0/24"
  map_public_ip_on_launch = true
  tags = [
    {
      key   = "name"
      value = "My-Subnet"
    }
  ]
}