//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccOpenshiftClustersDatasource(t *testing.T) {
	//// Due to Insufficient Capacity issues we can't run this test in a deterministic way
	//// And test is moro longer than runner CI can except
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
resource "numspot_openshift_cluster" "terraform_openshift_cluster_acctest" {
  cidr                   = "172.18.0.0/24"
  name                   = "Terraform"
  description            = "Terraform"
  availability_zone_name = "cloudgouv-eu-west-1a"

  node_pools = [
    {
      name         = "testnp1"
      node_count   = 2
      node_profile = "MEDIUM"
    }
  ]

  version = "4.17.15"
}

data "numspot_openshift_clusters" "datasource_openshift_clusters_acctest" {}

output "selected_cluster" {
  value = try([for c in data.numspot_openshift_clusters.datasource_openshift_clusters_acctest.items : c if c.id == numspot_openshift_cluster.terraform_openshift_cluster_acctest.id][0], null)
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.numspot_openshift_clusters.datasource_openshift_clusters_acctest", "items.*",
						map[string]string{
							"cidr":        "172.18.0.0/24",
							"name":        "Terraform",
							"description": "Terraform",
							"version":     "4.17.15",
						},
					),
				),
			},
		},
	})
}
