//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccKubernetesNodePoolResource(t *testing.T) {
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
  autoscaling = {
    max = 6
    min = 3
  }
  root_disk = {
    size = 30
    iops = 1000
    type = "enterprise"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"numspot_kubernetes_nodepool.terraform_kubernetes_np_acctest",
						"name",
						"terraform-kubernetes-np-acctest",
					),
					resource.TestCheckResourceAttr("numspot_kubernetes_nodepool.terraform_kubernetes_np_acctest", "availability_zone", "eu-west-2c"),
					resource.TestCheckResourceAttr("numspot_kubernetes_nodepool.terraform_kubernetes_np_acctest", "node_profile", "small"),
					resource.TestCheckResourceAttr("numspot_kubernetes_nodepool.terraform_kubernetes_np_acctest", "replicas", "3"),
				),
			},
			{
				ResourceName: "numspot_kubernetes_nodepool.terraform_kubernetes_np_acctest",
				ImportState:  true,
				ImportStateIdFunc: resource.ImportStateIdFunc(func(state *terraform.State) (string, error) {
					clusterId := state.RootModule().Resources["numspot_kubernetes_cluster.terraform_kubernetes_cluster_acctest"].Primary.ID
					nodepoolId := state.RootModule().Resources["numspot_kubernetes_nodepool.terraform_kubernetes_np_acctest"].Primary.ID
					return clusterId + "/" + nodepoolId, nil
				}),
				ImportStateVerify: true,
				// status/autoscaling are not returned by the nodepool read,
				// and replicas returns the *current* count (0 while the pool
				// is provisioning), so none can be verified after import.
				ImportStateVerifyIgnore: []string{"status", "autoscaling", "replicas"},
			},
		},
	})
}
