//go:build acc

package provider

import (
	b64 "encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

// This struct will store the input data that will be used in your tests (all fields as string)
type StepDataKeyPair struct {
	name,
	publicKey,
	id string
}

// Generate checks to validate that resource 'numspot_keypair.test' has input data values
func getFieldMatchChecksKeyPair(data StepDataKeyPair) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("numspot_keypair.test", "name", data.name),
		resource.TestCheckResourceAttr("numspot_keypair.test", "id", data.id),
		resource.TestCheckResourceAttr("numspot_keypair.test", "public_key", data.publicKey),
	}
}

func TestAccKeyPairResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	var resourceId string

	////////////// Define input data that will be used in the test sequence //////////////
	// resource fields that can be updated in-place	{{not updatable resource fields}}
	id := "the-id"
	idUpdated := "the-id-updated"

	// resource fields that cannot be updated in-place (requires replace)
	name := "key-pair-name-terraform"
	nameUpdated := "key-pair-name-terraform-updated"
	publicKey := b64.StdEncoding.EncodeToString([]byte("publicKey"))
	publicKeyUpdated := b64.StdEncoding.EncodeToString([]byte("publicKey-updated"))
	/////////////////////////////////////////////////////////////////////////////////////

	////////////// Define plan values and generate associated attribute checks  //////////////
	// The base plan (used in first create and to reset resource state before some tests)
	basePlanValues := StepDataKeyPair{
		name:      name,
		id:        id,
		publicKey: publicKey,
	}

	createChecks := append(
		getFieldMatchChecksKeyPair(basePlanValues),

		resource.TestCheckResourceAttrWith("numspot_keypair.test", "private_key", func(v string) error {
			require.NotEmpty(t, v)
			resourceId = v
			return nil
		}),
	)

	// The plan that should trigger Update function (based on basePlanValues). Update the value for as much updatable fields as possible here.
	updatePlanValues := StepDataKeyPair{
		name:      name, // Update values for updatable fields
		id:        idUpdated,
		publicKey: publicKey,
	}
	updateChecks := append(
		getFieldMatchChecksKeyPair(updatePlanValues),

		resource.TestCheckResourceAttrWith("numspot_keypair.test", "private_key", func(v string) error {
			require.NotEmpty(t, v)
			require.Equal(t, v, resourceId)
			return nil
		}),
	)

	// The plan that should trigger Replace behavior (based on basePlanValues or updatePlanValues). Update the value for as much non-updatable fields as possible here.
	replacePlanValues := StepDataKeyPair{
		name:      nameUpdated, // Update values for non-updatable fields
		id:        id,
		publicKey: publicKeyUpdated,
	}
	replaceChecks := append(
		getFieldMatchChecksKeyPair(replacePlanValues),

		resource.TestCheckResourceAttrWith("numspot_keypair.test", "private_key", func(v string) error {
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
				ImportStateVerifyIgnore: []string{"private_key"},
			},
			// Update testing Without Replace
			{
				Config: testKeyPairConfig(updatePlanValues),
				Check:  resource.ComposeAggregateTestCheckFunc(updateChecks...),
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
  id         = %[1]q
  name       = %[2]q
  public_key = %[3]q
}`, data.id, data.name, data.publicKey,
	)
}
