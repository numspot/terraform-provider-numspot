resource "numspot_postgres_cluster" "postgres-cluster" {
  name = "terraform-resource"
  user = "terraform"
  node_configuration = {
    vcpu_count        = 2
    performance_level = "MEDIUM"
    memory_size_gi_b  = 2
  }
  volume = {
    type      = "GP2",
    size_gi_b = 10
  }
  visibility = "INTERNAL"
  extensions = [{
    name    = "TIMESCALEDB"
    version = "string"
  }]
  replica_count = 1
}
