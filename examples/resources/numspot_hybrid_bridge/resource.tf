resource "numspot_vpc" "vpc" {
  ip_range = "10.101.1.0/24"
  tags = [{
    key   = "name"
    value = "terraform-hybrid-bridge-acctest"
  }]
}

resource "numspot_hybrid_bridge" "hybrid-bridge" {
  managed_service_id = ""
  vpc_id             = numspot_vpc.vpc.id
}