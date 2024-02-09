package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccNicResource(t *testing.T) {
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNicConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nic.subnet_id", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "field", func(v string) error {
						require.NotEmpty(t, v)
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_nic.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			{
				Config: testNicConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_nic.test", "field", "value"),
					resource.TestCheckResourceAttrWith("numspot_nic.test", "field", func(v string) error {
						return nil
					}),
				),
			},
		},
	})
}

func testNicConfig() string {
	return `
resource "numspot_net" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
	net_id 		= numspot_net.net.id
	ip_range 	= "10.101.1.0/24"
}


resource "numspot_nic" "test" {
	subnet_id = numspot_subnet.subnet.id
}
`
}
