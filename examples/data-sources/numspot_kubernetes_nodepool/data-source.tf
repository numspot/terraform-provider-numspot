resource "numspot_kubernetes_cluster" "kubernetes-cluster" {
  cidr       = "10.20.0.0/16"
  name       = "test-tf-kube"
  profile    = "small"
  version    = "1.32"
  visibility = "EXTERNAL"
}

resource "numspot_kubernetes_nodepool" "kubernetes-np" {
  cluster_id        = numspot_kubernetes_cluster.kubernetes-cluster.id
  name              = "test-tf-nodepool"
  availability_zone = "eu-west-2c"
  node_profile      = "small"
  replicas          = 3
  root_disk = {
    size = 31
    iops = 101
    type = "performance"
  }
  depends_on = [numspot_kubernetes_cluster.kubernetes-cluster]
}

data "numspot_kubernetes_nodepools" "datasource-nodepools" {
  cluster_id = numspot_kubernetes_cluster.kubernetes-cluster.id
  depends_on = [numspot_kubernetes_nodepool.kubernetes-np]
}
