//go:build acc

package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccBucketsDatasource(t *testing.T) {
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
resource "numspot_bucket" "bucket-acctest" {
  name = "bucket-acctest"
}

data "numspot_buckets" "datasource-buckets-acctest" {
  depends_on = [numspot_bucket.bucket-acctest]
}

output "selected_buckets" {
  value = try([for c in data.numspot_buckets.datasource-buckets-acctest.items : c if c.name == numspot_bucket.bucket-acctest.name][0], null)
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.numspot_buckets.datasource-buckets-acctest", "items.*",
						map[string]string{
							"name": "bucket-acctest",
						},
					),
				),
			},
		},
	})
}
