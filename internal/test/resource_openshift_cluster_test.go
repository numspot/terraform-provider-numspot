//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccOpenshiftClusterResource(t *testing.T) {
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
			// Step 1 - Create cluster
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
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_openshift_cluster.terraform_openshift_cluster_acctest", "cidr", "172.18.0.0/24"),
					resource.TestCheckResourceAttr("numspot_openshift_cluster.terraform_openshift_cluster_acctest", "name", "Terraform"),
					resource.TestCheckResourceAttr("numspot_openshift_cluster.terraform_openshift_cluster_acctest", "description", "Terraform"),
					resource.TestCheckResourceAttr("numspot_openshift_cluster.terraform_openshift_cluster_acctest", "version", "4.17.15"),
					resource.TestCheckResourceAttr("numspot_openshift_cluster.terraform_openshift_cluster_acctest", "nodePools.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"numspot_openshift_cluster.terraform_openshift_cluster_acctest",
						"nodePools.*",
						map[string]string{
							"name":         "testnp1",
							"node_count":   "2",
							"node_profile": "MEDIUM",
						},
					),
				),
			},
			// Step 2 - Import cluster
			{
				ResourceName:            "numspot_openshift_cluster.terraform_openshift_cluster_acctest",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
		},
	})
}
