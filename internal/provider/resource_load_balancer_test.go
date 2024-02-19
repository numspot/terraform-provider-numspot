package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type CreateLBTestData struct {
	name        string
	subnetID    string
	backendPort int
	lbPort      int
	lbProtocol  string
	lbType      string
}
type UpdateLBTestData struct {
	name                 string
	subnetID             string
	backendPort          int
	lbPort               int
	lbProtocol           string
	lbType               string
	hcCheckInterval      int
	hcHealthyThreshold   int
	hcPath               string
	hcPort               int
	hcProtocol           string
	hcTimeout            int
	hcUnhealthyThreshold int
}

func TestAccLoadBalancerResource(t *testing.T) {
	createLbData := CreateLBTestData{
		name:        "elb-test",
		subnetID:    "subnet-3709dfbf",
		backendPort: 80,
		lbPort:      80,
		lbProtocol:  "TCP",
		lbType:      "internal",
	}

	updateLbData := UpdateLBTestData{
		name:                 "elb-test",
		subnetID:             "subnet-3709dfbf",
		backendPort:          80,
		lbPort:               80,
		lbProtocol:           "TCP",
		lbType:               "internal",
		hcCheckInterval:      30,
		hcHealthyThreshold:   10,
		hcPath:               "/index.html",
		hcPort:               8080,
		hcProtocol:           "HTTPS",
		hcTimeout:            5,
		hcUnhealthyThreshold: 5,
	}

	pr := TestAccProtoV6ProviderFactories
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: testLoadBalancerConfig_Create(createLbData),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "name", createLbData.name),
				),
			},
			// ImportState testing
			{
				ResourceName:            "numspot_load_balancer.testlb",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			//Update testing
			{
				Config: testLoadBalancerConfig_Update(updateLbData),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.check_interval", strconv.Itoa(updateLbData.hcCheckInterval)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.healthy_threshold", strconv.Itoa(updateLbData.hcHealthyThreshold)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.path", updateLbData.hcPath),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.port", strconv.Itoa(updateLbData.hcPort)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.protocol", updateLbData.hcProtocol),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.timeout", strconv.Itoa(updateLbData.hcTimeout)),
					resource.TestCheckResourceAttr("numspot_load_balancer.testlb", "health_check.unhealthy_threshold", strconv.Itoa(updateLbData.hcUnhealthyThreshold)),
					//resource.TestCheckResourceAttrWith("numspot_load_balancer.testlb", "field", func(v string) error {
					//	return nil
					//}),
				),
			},
		},
	})
}

func testLoadBalancerConfig_Create(createLbData CreateLBTestData) string {
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
}`, createLbData.name,
		createLbData.backendPort,
		createLbData.lbPort,
		createLbData.lbProtocol,
		createLbData.subnetID,
		createLbData.lbType)
}

func testLoadBalancerConfig_Update(updateLbData UpdateLBTestData) string {
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
			health_check = {
				check_interval = %d
				healthy_threshold = %d
				path = "%s"
				port = %d
				protocol = "%s"
				timeout = %d
				unhealthy_threshold = %d
			}
}`, updateLbData.name,
		updateLbData.backendPort,
		updateLbData.lbPort,
		updateLbData.lbProtocol,
		updateLbData.subnetID,
		updateLbData.lbType,
		updateLbData.hcCheckInterval,
		updateLbData.hcHealthyThreshold,
		updateLbData.hcPath,
		updateLbData.hcPort,
		updateLbData.hcProtocol,
		updateLbData.hcTimeout,
		updateLbData.hcUnhealthyThreshold)
}
