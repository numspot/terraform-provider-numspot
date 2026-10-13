resource "numspot_postgres_cluster" "postgres-cluster" {
  name = "terraform-resource"
  user = "terraform"
  node_configuration = {
    vcpu_count       = 2
    memory_size_gi_b = 2
  }
  volume = {
    type      = "ENTERPRISE"
    size_gi_b = 10
  }
  visibility = "EXTERNAL"
  extensions = [{
    name = "TIMESCALEDB"
  }]
  replica_count = 1
  major_version = "18"
}

data "numspot_postgres_clusters" "postgres-clusters" {
  depends_on = [numspot_postgres_cluster.postgres-cluster]
}
