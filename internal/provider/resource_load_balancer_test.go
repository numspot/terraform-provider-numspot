package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoadBalancerResource(t *testing.T) {
	createLbName := "elb-test"
	createSubnetID := "subnet-3709dfbf" //with name test_tf_lb on cockpit and attached to vpc vpc-be5c2155
	createBackendPort := 80
	createLbPort := 80
	createLbProtocol := "TCP"
	createLbType := "internal"
	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testLoadBalancerConfig_Create(
					createLbName,
					createSubnetID,
					createLbProtocol,
					createLbType,
					createBackendPort,
					createLbPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", createLbName),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_load_balancer.testlb",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Update testing
			//{
			//	Config: testLoadBalancerConfig_Update(),
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("numspot_load_balancer.test", "field", "value"),
			//		resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "field", func(v string) error {
			//			return nil
			//		}),
			//	),
			//},
		},
	})
}

func testLoadBalancerConfig_Create(lbName, subnetID, lbProtocol, lbtype string, backendPort, lbPort int) string {
	return fmt.Sprintf(`resource "numspot_load_balancer" "testlb" {
			name = "%s"
			listeners = [
				{
					backend_port = %d
					load_balancer_port = %d
					load_balancer_protocol = "%s"
					
				}
			]
			subnets = ["%s"]
			type = "%s"
}`, lbName, backendPort, lbPort, lbProtocol, subnetID, lbtype)
}

func testLoadBalancerConfig_Update() string {
	return `resource "numspot_load_balancer" "test" {
    			}`
}
