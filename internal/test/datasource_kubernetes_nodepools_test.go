//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccKubernetesNodePoolsDatasource(t *testing.T) {
	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: `
resource "numspot_kubernetes_cluster" "terraform_kubernetes_cluster_acctest" {
  cidr       = "10.10.0.0/16"
  name       = "terraform-kubernetes-cluster-test"
  profile    = "xsmall"
  version    = "1.33"
  visibility = "EXTERNAL"
}

resource "numspot_kubernetes_nodepool" "terraform_kubernetes_np_acctest" {
  cluster_id        = numspot_kubernetes_cluster.terraform_kubernetes_cluster_acctest.id
  name              = "terraform-kubernetes-np-acctest"
  availability_zone = "eu-west-2c"
  node_profile      = "small"
  replicas          = 3
  root_disk = {
    size = 30
    iops = 1000
    type = "enterprise"
  }
}

data "numspot_kubernetes_nodepools" "datasource_kubernetes_nodepools_acctest" {
  cluster_id = numspot_kubernetes_cluster.terraform_kubernetes_cluster_acctest.id
  depends_on = [numspot_kubernetes_nodepool.terraform_kubernetes_np_acctest]
}
`, Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"cluster_id",
						"numspot_kubernetes_cluster.terraform_kubernetes_cluster_acctest",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.#",
						"1",
					),
					resource.TestCheckResourceAttrSet(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.id",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.name",
						"terraform-kubernetes-np-acctest",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.availability_zone",
						"eu-west-2c",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.node_profile",
						"small",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.replicas",
						"3",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.root_disk.size",
						"30",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.root_disk.type",
						"enterprise",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.root_disk.iops",
						"1000",
					),
					resource.TestCheckResourceAttr(
						"data.numspot_kubernetes_nodepools.datasource_kubernetes_nodepools_acctest",
						"items.0.root_disk.type",
						"enterprise",
					),
				),
			},
		},
	})
}
