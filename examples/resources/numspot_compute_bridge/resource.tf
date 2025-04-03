resource "numspot_vpc" "vpc-source" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "vpc-dest" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_compute_bridge" "compute-bridge" {
  source_vpc_id      = numspot_vpc.vpc-source.id
  destination_vpc_id = numspot_vpc.vpc-dest.id
}