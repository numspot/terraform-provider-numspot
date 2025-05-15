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