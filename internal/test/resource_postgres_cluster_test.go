//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccPostgresClusterResource(t *testing.T) {
	// Waiting account e2e with hcp
	t.Skip()
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
resource "numspot_postgres_cluster" "terraform_postgres_cluster_acctest" {
  name = "terraform_postgres_cluster_acctest"
  user = "team.terraform"
  node_configuration = {
    vcpu_count        = 2
    performance_level = "MEDIUM"
    memory_size_gi_b  = 2
  }
  volume = {
    type      = "IO1",
    size_gi_b = 10
  }
  is_public = false
  extensions = [{
    name    = "timescaledb"
    version = "string"
  }]
  replica_count = 1
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"numspot_postgres_cluster.terraform_postgres_cluster_acctest",
						"name",
						"terraform_postgres_cluster_acctest",
					),
					resource.TestCheckResourceAttr("numspot_postgres_cluster.terraform_postgres_cluster_acctest", "user", "team.terraform"),
					resource.TestCheckResourceAttr("numspot_postgres_cluster.terraform_postgres_cluster_acctest", "is_public", "false"),
					resource.TestCheckResourceAttr(
						"numspot_postgres_cluster.terraform_postgres_cluster_acctest",
						"node_configuration.vcpu_count",
						"2",
					),
					resource.TestCheckResourceAttr(
						"numspot_postgres_cluster.terraform_postgres_cluster_acctest",
						"node_configuration.memory_size_gi_b",
						"2",
					),
					resource.TestCheckResourceAttr("numspot_postgres_cluster.terraform_postgres_cluster_acctest", "volume.type", "IO1"),
					resource.TestCheckResourceAttr("numspot_postgres_cluster.terraform_postgres_cluster_acctest", "volume.size_gi_b", "10"),
					resource.TestCheckResourceAttr("numspot_postgres_cluster.terraform_postgres_cluster_acctest", "replica_count", "1"),
				),
			},
		},
	})
}
