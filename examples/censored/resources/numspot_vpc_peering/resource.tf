resource "numspot_vpc" "source" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "accepter" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter.id
  source_vpc_id   = numspot_vpc.source.id
}