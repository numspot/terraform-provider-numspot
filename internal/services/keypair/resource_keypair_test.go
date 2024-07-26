//go:build acc

package keypair_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataKeyPair struct {
	name,
	publicKey string
}

// Generate checks to validate that resource 'numspot_keypair.test' has input data values
func getFieldMatchChecksKeyPair(data StepDataKeyPair) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_keypair.test", "name", data.name),
		resource.TestCheckResourceAttr("numspot_keypair.test", "public_key", data.publicKey),
	}
}

func TestAccKeyPairResource(t *testing.T) {
	pr := provider.TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place	{{not updatable resource fields}}
	// None

	// resource fields that cannot be updated in-place (requires replace)
	name := "key-pair-name-terraform"
	nameUpdated := "key-pair-name-terraform-updated"
	publicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEm78d7vfikcOXDdvT0yioYUDm3spxjVws/xnL0J5f0P"
	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataKeyPair{
		name:      name,
		publicKey: publicKey,
	}

	createChecks := append(
		getFieldMatchChecksKeyPair(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_keypair.test", "name", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataKeyPair{
		name:      nameUpdated, // Update values for non-updatable fields
		publicKey: publicKey,
	}
	replaceChecks := append(
		getFieldMatchChecksKeyPair(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_keypair.test", "name", func(v string) error {
			require.NotEmpty(t, v)
			require.NotEqual(t, v, resourceId)
			return nil
		}),
	)
	/////////////////////////////////////////////////////////////////////////////////////

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // Create testing
				Config: testKeyPairConfig(basePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(createChecks...),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_keypair.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{""},
			},
			// Update testing With Replace
			{
				Config: testKeyPairConfig(replacePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(replaceChecks...),
			},
		},
	})
}

func testKeyPairConfig(data StepDataKeyPair) string {
	return fmt.Sprintf(`
resource "numspot_keypair" "test" {
  name       = %[1]q
  public_key = %[2]q
}`, data.name, data.publicKey,
	)
}
