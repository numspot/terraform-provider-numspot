resource "numspot_vpc" "vpc_source" {
  ip_range = "10.101.1.0/24"
  tags = [{
    key   = "name"
    value = "vpc a"
  }]
}

resource "numspot_vpc" "vpc_dest" {
  ip_range = "10.101.2.0/24"
  tags = [{
    key   = "name"
    value = "vpc b"
  }]
}

resource "numspot_compute_bridge" "compute-bridge" {
  destination_vpc_id = numspot_vpc.vpc_dest.id
  source_vpc_id      = numspot_vpc.vpc_source.id
}

data "numspot_compute_bridges" "datasource-compute-bridge" {
  depends_on = [numspot_compute_bridge.compute-bridge]
}