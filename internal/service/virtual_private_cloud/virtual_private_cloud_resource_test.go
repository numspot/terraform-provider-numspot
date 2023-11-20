package virtual_private_cloud_test

import (
	"fmt"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"numspot": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestVPCResource_Create(t *testing.T) {
	ipRange := "172.16.0.0/16"
	tenancy := "dedicated"
	tenancyUpdate := "default"

	var firstVpcId string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testVPCConfig_Create(ipRange, tenancy),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.main", "ip_range", ipRange),
					resource.TestCheckResourceAttr("numspot_vpc.main", "tenancy", tenancy),
					resource.TestCheckResourceAttrWith("numspot_vpc.main", "id", func(v string) error {
						require.NotEmpty(t, v)
						firstVpcId = v
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:      "numspot_vpc.main",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testVPCConfig_Create(ipRange, tenancyUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc.main", "ip_range", ipRange),
					resource.TestCheckResourceAttr("numspot_vpc.main", "tenancy", tenancyUpdate),
					resource.TestCheckResourceAttrWith("numspot_vpc.main", "id", func(v string) error {
						require.NotEmpty(t, v)
						require.NotEqual(t, firstVpcId, v) // Ensure VPC ID changed --> Replaced
						return nil
					}),
				),
			},
		},
	})
}

func testVPCConfig_Create(ipRange, tenancy string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "main" {
	ip_range 		= %[1]q
	tenancy			= %[2]q
}
`, ipRange, tenancy)
}
