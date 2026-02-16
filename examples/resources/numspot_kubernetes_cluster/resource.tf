resource "numspot_kubernetes_cluster" "kubernetes-cluster" {
  cidr       = "10.20.0.0/16"
  name       = "test-tf-kube"
  profile    = "small"
  version    = "1.32"
  visibility = "EXTERNAL"
}
