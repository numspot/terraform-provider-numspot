//go:build acc

package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccPostgresClustersDatasource(t *testing.T) {
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
    iops      = 200,
    size_gi_b = 10
  }
  is_public = false
  extensions = [{
    name    = "timescaledb"
    version = "string"
  }]
  replica_count = 1
}

data "numspot_postgres_clusters" "datasource_postgres_clusters_acctest" {
  depends_on = [numspot_postgres_cluster.terraform_postgres_cluster_acctest]
}

output "selected_cluster" {
  value = try([for c in data.numspot_postgres_clusters.datasource_postgres_clusters_acctest.items : c if c.id == numspot_postgres_cluster.terraform_postgres_cluster_acctest.id][0], null)
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckFunc(func(s *terraform.State) error {
						ds, ok := s.RootModule().Resources["data.numspot_postgres_clusters.datasource_postgres_clusters_acctest"]
						if !ok {
							return fmt.Errorf("datasource not found")
						}

						rs, ok := s.RootModule().Resources["numspot_postgres_cluster.terraform_postgres_cluster_acctest"]
						if !ok {
							return fmt.Errorf("postgres cluster resource not found")
						}

						targetID := rs.Primary.ID

						for k, v := range ds.Primary.Attributes {
							if strings.HasSuffix(k, ".id") && v == targetID {
								return nil
							}
						}

						return fmt.Errorf("postgres cluster with id %s not found in datasource items", targetID)
					}),

					resource.TestCheckFunc(func(s *terraform.State) error {
						ds := s.RootModule().Resources["data.numspot_postgres_clusters.datasource_postgres_clusters_acctest"]
						rs := s.RootModule().Resources["numspot_postgres_cluster.terraform_postgres_cluster_acctest"]

						targetID := rs.Primary.ID

						for k, v := range ds.Primary.Attributes {
							if strings.HasSuffix(k, ".id") && v == targetID {
								prefix := strings.TrimSuffix(k, ".id")

								if ds.Primary.Attributes[prefix+".name"] != "terraform_postgres_cluster_acctest" {
									return fmt.Errorf("name mismatch")
								}
								if ds.Primary.Attributes[prefix+".state"] != "ACTIVE" {
									return fmt.Errorf("state mismatch")
								}
								if ds.Primary.Attributes[prefix+".user"] != "team.terraform" {
									return fmt.Errorf("user mismatch")
								}

								return nil
							}
						}

						return fmt.Errorf("postgres cluster %s not found", targetID)
					}),
				),
			},
		},
	})
}
