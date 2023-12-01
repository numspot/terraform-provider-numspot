package internet_gateway_test

import (
	"fmt"
	"testing"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestInternetGatewayResourceCreate(t *testing.T) {
	firstVPCID := "vpc-c3726ca8"
	secondVPCID := "vpc-1c53cc9c"

	var internetGatewayID string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testInternetGatewayConfigCreate(firstVPCID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_internet_gateway.main", "virtual_private_cloud_id", firstVPCID),
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.main", "id", func(v string) error {
						require.NotEmpty(t, v)
						internetGatewayID = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_internet_gateway.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testInternetGatewayConfigCreate(secondVPCID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_internet_gateway.main", "virtual_private_cloud_id", secondVPCID),
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.main", "id", func(secondInternetGatewayID string) error {
						require.Equal(t, internetGatewayID, secondInternetGatewayID)
						return nil
					}),
				),
			},
			// Create without vpcID testing
			{
				Config: fmt.Sprintf(`resource "numspot_internet_gateway" "empty" {}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("numspot_internet_gateway.empty", "id", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
		},
	})
}

func testInternetGatewayConfigCreate(virtualPrivateCloudId string) string {
	return fmt.Sprintf(`
resource "numspot_internet_gateway" "main" {
	virtual_private_cloud_id = %[1]q
}
`, virtualPrivateCloudId)
}
