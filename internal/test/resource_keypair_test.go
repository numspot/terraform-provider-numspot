package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccKeyPairResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create testing
				Config: `
resource "numspot_keypair" "test" {
  name       = "key-pair-name-terraform"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEm78d7vfikcOXDdvT0yioYUDm3spxjVws/xnL0J5f0P"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_keypair.test", "name", "key-pair-name-terraform"),
					resource.TestCheckResourceAttr("numspot_keypair.test", "public_key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEm78d7vfikcOXDdvT0yioYUDm3spxjVws/xnL0J5f0P"),
					resource.TestCheckResourceAttrWith("numspot_keypair.test", "name", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
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
			// 3 - Update testing With Replace
			{
				Config: `
resource "numspot_keypair" "test" {
  name       = "key-pair-name-terraform-updated"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEm78d7vfikcOXDdvT0yioYUDm3spxjVws/xnL0J5f0P"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_keypair.test", "name", "key-pair-name-terraform-updated"),
					resource.TestCheckResourceAttr("numspot_keypair.test", "public_key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEm78d7vfikcOXDdvT0yioYUDm3spxjVws/xnL0J5f0P"),
					resource.TestCheckResourceAttrWith("numspot_keypair.test", "name", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						return nil
					}),
				),
			},
		},
	})
}
