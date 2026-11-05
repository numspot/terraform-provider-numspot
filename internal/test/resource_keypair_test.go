//go:build acc

package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

// Test cases
//
// Create Keypair with public key
// Replace Keypair -> new keypair without public key
// Recreate Keypair with same name

func TestAccKeyPairResource(t *testing.T) {
	publicKey := acctest.GetEnvOrReplay(t, "KEYPAIR_PUBLIC_KEY_ACCTEST", "KEYPAIR_PUBLIC_KEY")

	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create Keypair with public key
				Config: fmt.Sprintf(`
resource "numspot_keypair" "test" {
  name       = "key-pair-name-terraform"
  public_key = "%s"
}`, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_keypair.test", "name", "key-pair-name-terraform"),
					resource.TestCheckResourceAttr("numspot_keypair.test", "public_key", publicKey),
					resource.TestCheckResourceAttrWith("numspot_keypair.test", "name", func(v string) error {
						return acctest.InitResourceId(t, v, &resourceId)
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_keypair.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{""},
			},
			// 3 - Replace keypair => no public key
			{
				Config: `
resource "numspot_keypair" "test" {
  name = "key-pair-name-terraform-updated"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_keypair.test", "name", "key-pair-name-terraform-updated"),
					resource.TestCheckResourceAttrWith("numspot_keypair.test", "name", func(v string) error {
						return acctest.CheckResourceIdChanged(t, v, &resourceId)
					}),
				),
			},
			// 4 - Recreate keypair with same name
			{
				Config: `
resource "numspot_keypair" "test_recreated" {
  name = "key-pair-name-terraform-updated"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_keypair.test_recreated", "name", "key-pair-name-terraform-updated"),
					resource.TestCheckResourceAttrWith("numspot_keypair.test_recreated", "name", func(v string) error {
						return acctest.CheckResourceIdUnchanged(t, v, &resourceId)
					}),
				),
			},
		},
	})
}
