resource "numspot_kubernetes_cluster" "kubernetes-cluster" {
  cidr       = "10.20.0.0/16"
  name       = "test-tf-kube"
  profile    = "small"
  version    = "1.33"
  visibility = "EXTERNAL"
}

resource "numspot_kubernetes_nodepool" "kubernetes-np" {
  cluster_id        = numspot_kubernetes_cluster.kubernetes-cluster.id
  name              = "terraform-kubernetes-np-test"
  availability_zone = "eu-west-2c"
  node_profile      = "small"
  replicas          = 3
  autoscaling = {
    min = 1
    max = 1
  }
  root_disk = {
    size = 30
    iops = 1000
    type = "enterprise"
  }
  depends_on = [numspot_kubernetes_cluster.kubernetes-cluster]
}
