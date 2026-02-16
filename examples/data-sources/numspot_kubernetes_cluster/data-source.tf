resource "numspot_kubernetes_cluster" "kubernetes-cluster" {
  cidr       = "10.20.0.0/16"
  name       = "test-tf-kube"
  profile    = "small"
  version    = "1.32"
  visibility = "EXTERNAL"
}
data "numspot_kubernetes_clusters" "datasource-clusters" {
  depends_on = [numspot_kubernetes_cluster.kubernetes-cluster]
}
