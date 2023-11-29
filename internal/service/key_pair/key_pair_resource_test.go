package key_pair_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestKeypairResource_Create(t *testing.T) {
	name := fmt.Sprintf("%s%d", "TEST_TERRAFORM_KEYPAIR_", rand.Uint64())

	pr := acctest.TestAccProtoV6ProviderFactories

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,

		Steps: []resource.TestStep{
			{
				Config: testKeyPairConfig_Create(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_key_pair.create", "name", name),
					resource.TestCheckResourceAttr("numspot_key_pair.create", "id", name),
					resource.TestCheckResourceAttrWith("numspot_key_pair.create", "private_key", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_key_pair.create",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"private_key", "fingerprint"},
			},
			// Update testing
			{
				Config: testKeyPairConfig_Create(name + "_UPDATED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_key_pair.create", "name", name+"_UPDATED"),
					resource.TestCheckResourceAttr("numspot_key_pair.create", "id", name+"_UPDATED"),
					resource.TestCheckResourceAttrWith("numspot_key_pair.create", "private_key", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func TestKeypairResource_Import(t *testing.T) {
	name := fmt.Sprintf("%s%d", "TEST_TERRAFORM_KEYPAIR_", rand.Uint64())
	publicKey := "c3NoLXJzYSBBQUFBQjNOemFDMXljMkVBQUFBREFRQUJBQUFCQVFDT3dNUEorWjR4WUNXV2VKVWpk\nd1M3TC9JYzdEQ0RwUTZmU1pqekx5SktWTno4OFVqbm0yK09JYjVMcGVFMHhHZ1hLTWZlZ2hGYytl\nNlNVTnRCOUFwSGMrMlNub1Y3SjVkTktkb3FPTFpaRFVKR1EzTnVrNGtyS00zMWRyMlBSS2IzNHdV\nYzZYQUV6NXFQTUI2YTNvUm90eGxWWTduU1FEL0J2UUZXRkhQaHNiOWtsbkhTd2gwOHlIU3B0Q3hS\nUmhyQlc4Wmk1L25FY1hOTjlGTEZzaVNJWGliRVNJQWZrYlZMdFlUSUhOL2ZWTWpqNVd5UjFRd3Rk\nMmUwb1RDZ1FhZUxZSWNhN0NjSmt3N0JvYXhKUlZRSmE5dm1LbUdSQjRoWDlpZTNDK2hjeEswRS9H\nVjN4dktwem91Vno0WEdSaktldEsvZmo1aWVjbm5yRnFOQXR2Z2ogdG90YXJAbG9jYWxob3N0Cg=="

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testKeyPairConfig_Import(name, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_key_pair.import", "name", name),
					resource.TestCheckResourceAttr("numspot_key_pair.import", "id", name),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_key_pair.import",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"private_key", "public_key"},
			},
			// Update testing
			{
				Config: testKeyPairConfig_Import(name+"_UPDATED", publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_key_pair.import", "name", name+"_UPDATED"),
					resource.TestCheckResourceAttr("numspot_key_pair.import", "id", name+"_UPDATED"),
				),
			},
		},
	})
}

func testKeyPairConfig_Create(name string) string {
	return fmt.Sprintf(`
resource "numspot_key_pair" "create" {
  name   		= %[1]q
}
`, name)
}

func testKeyPairConfig_Import(name, publicKey string) string {
	return fmt.Sprintf(`
resource "numspot_key_pair" "import" {
  name   		= %[1]q
  public_key 	= %[2]q
}
`, name, publicKey)
}
