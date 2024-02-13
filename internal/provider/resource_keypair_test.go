package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccKeyPairResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories

	randName := rand.Intn(9999-1000) + 1000
	name := fmt.Sprintf("key-pair-name-%d", randName)
	updatedName := fmt.Sprintf("updated-key-pair-name-%d", randName)
	privateKey := ""

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testKeyPairConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_keypair.test", "name", name),
					resource.TestCheckResourceAttr("numspot_keypair.test", "id", name),
					resource.TestCheckResourceAttrWith("numspot_keypair.test", "private_key", func(v string) error {
						require.NotEmpty(t, v)
						privateKey = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_keypair.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
			// Update testing
			{
				Config: testKeyPairConfig(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_keypair.test", "name", updatedName),
					resource.TestCheckResourceAttr("numspot_keypair.test", "id", updatedName),
					resource.TestCheckResourceAttrWith("numspot_keypair.test", "private_key", func(v string) error {
						require.NotEmpty(t, v)
						require.NotEqual(t, v, privateKey)
						return nil
					}),
				),
			},
		},
	})
}

func testKeyPairConfig(name string) string {
	return fmt.Sprintf(`
resource "numspot_keypair" "test" {
	name = %[1]q
}`, name)
}
