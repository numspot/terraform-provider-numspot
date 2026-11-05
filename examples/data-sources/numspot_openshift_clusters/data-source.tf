# resource "numspot_openshift_cluster" "terraform_openshift_cluster" {
#   cidr                   = "172.18.0.0/24"
#   name                   = "Name"
#   description            = "Description"
#   availability_zone_name = "eu-west-2a"
#
#   node_pools = [
#     {
#       name         = "testnp1"
#       node_count   = 2
#       node_profile = "MEDIUM"
#     }
#   ]
#
#   version = "4.17.0"
# }
#
# data "numspot_openshift_clusters" "datasource_openshift_clusters" {}
#
# output "selected_cluster" {
#   value = try([for c in data.numspot_openshift_clusters.datasource_openshift_clusters.items : c if c.id == numspot_openshift_cluster.terraform_openshift_cluster.id][0], null)
# }
