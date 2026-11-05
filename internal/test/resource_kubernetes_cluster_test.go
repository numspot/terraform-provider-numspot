//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccKubernetesClusterResource(t *testing.T) {
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
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_kubernetes_cluster.test", "cidr", "10.10.0.0/16"),
					resource.TestCheckResourceAttr("numspot_kubernetes_cluster.test", "name", "terraform-kubernetes-cluster-test"),
					resource.TestCheckResourceAttr("numspot_kubernetes_cluster.test", "profile", "xsmall"),
					resource.TestCheckResourceAttr("numspot_kubernetes_cluster.test", "version", "1.33"),
					resource.TestCheckResourceAttr("numspot_kubernetes_cluster.test", "visibility", "EXTERNAL"),
				),
			},
			{
				ResourceName:            "numspot_kubernetes_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts", "air_gap"},
			},
		},
	})
}
