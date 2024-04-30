//go:build acc

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetPeeringResource(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetPeeringConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("numspot_vpc_peering.test", "field", "value"),
				//resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "field", func(v string) error {
				//	require.NotEmpty(t, v)
				//	return nil
				//}),
				),
			},
			// ImportState testing
			// Update testing
			//{
			//	Config: testNetPeeringConfig(),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("numspot_vpc_peering.test", "field", "value"),
			//		resource.TestCheckResourceAttrWith("numspot_vpc_peering.test", "field", func(v string) error {
			//			return nil
			//		}),
			//	),
			//},
		},
	})
}

func testNetPeeringConfig() string {
	return `
resource "numspot_vpc" "source" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "accepter" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter.id
  source_vpc_id   = numspot_vpc.source.id
}`
}

func TestAccNetPeeringResource_Tags(t *testing.T) {
	t.Parallel()
	pr := TestAccProtoV6ProviderFactories

	tagKey := "name"
	tagValue := "Terraform-Test-Vpc-Peering"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testNetPeeringConfig_Tags(tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.#", "1"),
				),
			},
			// ImportState testing
			// Update testing
			{
				Config: testNetPeeringConfig_Tags(tagKey, tagValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.0.key", tagKey),
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.0.value", tagValue),
					resource.TestCheckResourceAttr("numspot_vpc_peering.test", "tags.#", "1"),
				),
			},
		},
	})
}

func testNetPeeringConfig_Tags(key, value string) string {
	return fmt.Sprintf(`
resource "numspot_vpc" "source" {
  ip_range = "10.101.1.0/24"
}

resource "numspot_vpc" "accepter" {
  ip_range = "10.101.2.0/24"
}

resource "numspot_vpc_peering" "test" {
  accepter_vpc_id = numspot_vpc.accepter.id
  source_vpc_id   = numspot_vpc.source.id

  tags = [
    {
      key   = %[1]q
      value = %[2]q
    }
  ]
}`, key, value)
}
