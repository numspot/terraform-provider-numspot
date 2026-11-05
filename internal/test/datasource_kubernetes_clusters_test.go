//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccKubernetesClustersDatasource(t *testing.T) {
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
resource "numspot_kubernetes_cluster" "test" {
  cidr       = "10.10.0.0/16"
  name       = "terraform-kubernetes-cluster-test"
  profile    = "xsmall"
  version    = "1.33"
  visibility = "EXTERNAL"
}

data "numspot_kubernetes_clusters" "test" {
  depends_on = [numspot_kubernetes_cluster.test]
}

output "selected_cluster" {
  value = try([for c in data.numspot_kubernetes_clusters.test.items : c if c.id == numspot_kubernetes_cluster.test.id][0], null)
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.numspot_kubernetes_clusters.test", "items.0.cidr", "10.10.0.0/16"),
					resource.TestCheckResourceAttr("data.numspot_kubernetes_clusters.test", "items.0.name", "terraform-kubernetes-cluster-test"),
					resource.TestCheckResourceAttr("data.numspot_kubernetes_clusters.test", "items.0.profile", "xsmall"),
					resource.TestCheckResourceAttr("data.numspot_kubernetes_clusters.test", "items.0.version", "1.33"),
					resource.TestCheckResourceAttr("data.numspot_kubernetes_clusters.test", "items.0.visibility", "EXTERNAL"),
				),
			},
		},
	})
}
