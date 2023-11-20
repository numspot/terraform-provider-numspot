package subnet_test

import (
	"fmt"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"numspot": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestSubnetResourceCreate(t *testing.T) {
	ipRange := "172.16.0.0/20"
	ipRangeUpdate := "172.16.16.0/20"
	virtualPrivateCloudId := "vpc-64669a51"

	var subnetId string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testSubnetConfigCreate(ipRange, virtualPrivateCloudId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.main", "ip_range", ipRange),
					resource.TestCheckResourceAttr("numspot_subnet.main", "virtual_private_cloud_id", virtualPrivateCloudId),
					resource.TestCheckResourceAttrWith("numspot_subnet.main", "id", func(v string) error {
						require.NotEmpty(t, v)
						subnetId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_subnet.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testSubnetConfigCreate(ipRangeUpdate, virtualPrivateCloudId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_subnet.main", "ip_range", ipRangeUpdate),
					resource.TestCheckResourceAttr("numspot_subnet.main", "virtual_private_cloud_id", virtualPrivateCloudId),
					resource.TestCheckResourceAttrWith("numspot_subnet.main", "id", func(v string) error {
						require.NotEmpty(t, v)
						require.NotEqual(t, subnetId, v)
						return nil
					}),
				),
			},
		},
	})
}

func testSubnetConfigCreate(ipRange, virtualPrivateCloudId string) string {
	return fmt.Sprintf(`
resource "numspot_subnet" "main" {
	ip_range = %[1]q
	virtual_private_cloud_id = %[2]q
}
`, ipRange, virtualPrivateCloudId)
}
