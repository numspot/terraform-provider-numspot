resource "numspot_vpc" "vpc" {
  ip_range = "10.101.1.0/24"
  tags = [{
    key   = "name"
    value = "vpc a"
  }]
}

resource "numspot_hybrid_bridge" "hybrid-bridge" {
  managed_service_id = "" // Managed service ID
  vpc_id             = numspot_vpc.vpc.id
}

data "numspot_hybrid_bridges" "datasource-hybrid-bridge" {
  depends_on = [numspot_hybrid_bridge.hybrid-bridge]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_hybrid_bridges.datasource-hybrid-bridge.items.0.id"
  }
}