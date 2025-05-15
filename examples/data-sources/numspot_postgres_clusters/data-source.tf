resource "numspot_postgres_cluster" "postgres-cluster" {
  name = "terraform-resource"
  user = "team.terraform"
  node_configuration = {
    vcpu_count        = 2
    performance_level = "MEDIUM"
    memory_size_gi_b  = 2
  }
  automatic_backup = false
  volume = {
    type      = "IO1",
    iops      = 200,
    size_gi_b = 10
  }
  allowed_ip_ranges = ["0.0.0.0/0"]
}

data "numspot_postgres_clusters" "postgres-clusters" {
  depends_on = [numspot_postgres_cluster.postgres-cluster]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_postgres_clusters.postgres-clusters.items.0.id"
  }
}