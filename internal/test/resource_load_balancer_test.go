package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

// This test sequence is quite long (~8 minutes)

func TestAccLoadBalancerResource(t *testing.T) {
	// We spotted a bug on Outscale side:
	// When a load balancer is created, deleted and the recreated with the same name, the API returns 409 Resource conflict
	// Bug reported to Outscale: https://support.outscale.com/hc/fr-fr/requests/378437
	// Setting random name is not compliant with recorded test cassettes
	// For instance We skip this test in the CI pipeline until Outscale fixes the bug or we foind a better solution
	t.Skip()
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	var resourceId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{ // 1 - Create testing
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_vm" "test" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-2c4r"

  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_load_balancer" "test" {
  name = "elb-terraform-test"
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    }
  ]

  subnets         = [numspot_subnet.test.id]
  security_groups = [numspot_security_group.test.id]
  backend_vm_ids  = [numspot_vm.test.id]

  type = "internal"

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume",
					}),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "listeners.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "80",
						"load_balancer_port": "80",
					}),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "subnets.*", "numspot_subnet.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "security_groups.*", "numspot_security_group.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "backend_vm_ids.*", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "numspot_load_balancer.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// 3 - Update testing Without Replace
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },

  ]
}

resource "numspot_vm" "test" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-4c8r"

  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_load_balancer" "test" {
  name = "elb-terraform-test"
  listeners = [
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 8080
      load_balancer_port     = 8080
      load_balancer_protocol = "TCP"
    }
  ]

  subnets         = [numspot_subnet.test.id]
  security_groups = [numspot_security_group.test.id]
  backend_vm_ids  = [numspot_vm.test.id]

  type = "internal"

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "443",
						"load_balancer_port": "443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "8080",
						"load_balancer_port": "8080",
					}),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "subnets.*", "numspot_subnet.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "security_groups.*", "numspot_security_group.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "backend_vm_ids.*", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged. Expected %s but got %s.", resourceId, v)
						}
						return nil
					}),
				),
			},
			// 4 - Update testing With Replace (if needed)
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },

  ]
}

resource "numspot_vm" "test" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-4c8r"

  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_load_balancer" "test" {
  name = "elb-terraform-test-updated"
  listeners = [
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 8080
      load_balancer_port     = 8080
      load_balancer_protocol = "TCP"
    }
  ]

  subnets         = [numspot_subnet.test.id]
  security_groups = [numspot_security_group.test.id]
  backend_vm_ids  = [numspot_vm.test.id]

  type = "internal"

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", "elb-terraform-test-updated"),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "443",
						"load_balancer_port": "443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "8080",
						"load_balancer_port": "8080",
					}),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "subnets.*", "numspot_subnet.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "security_groups.*", "numspot_security_group.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "backend_vm_ids.*", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 5 - Update from Internal Load Balancer to Public Load Balancer
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_internet_gateway" "ig" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_subnet" "test" {
  vpc_id                  = numspot_vpc.vpc.id
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = true
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.vpc.id
  subnet_id = numspot_subnet.test.id

  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.ig.id
    }
  ]
}

resource "numspot_vm" "test" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-4c8r"

  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_load_balancer" "test" {
  name = "elb-terraform-test-updated"
  listeners = [
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 8080
      load_balancer_port     = 8080
      load_balancer_protocol = "TCP"
    }
  ]

  subnets         = [numspot_subnet.test.id]
  security_groups = [numspot_security_group.test.id]
  backend_vm_ids  = [numspot_vm.test.id]

  type = "internet-facing"

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]

  depends_on = [numspot_internet_gateway.ig]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", "elb-terraform-test-updated"),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "type", "internet-facing"),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "443",
						"load_balancer_port": "443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "8080",
						"load_balancer_port": "8080",
					}),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "subnets.*", "numspot_subnet.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "security_groups.*", "numspot_security_group.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "backend_vm_ids.*", "numspot_vm.test", "id"),
					resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// 6 - Update testing With Replace of dependency resource and with Replace of the resource (if needed)
			// This test is useful to check wether or not the deletion of the dependencies and then the deletion of the main resource works properly
			{
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_subnet" "test_new" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test_new" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },

  ]
}

resource "numspot_vm" "test_new" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-4c8r"

  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_load_balancer" "test" {
  name = "elb-terraform-test-updated"
  listeners = [
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 8080
      load_balancer_port     = 8080
      load_balancer_protocol = "TCP"
    }
  ]

  subnets         = [numspot_subnet.test_new.id]
  security_groups = [numspot_security_group.test_new.id]
  backend_vm_ids  = [numspot_vm.test_new.id]

  type = "internal"

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", "elb-terraform-test-updated"),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "443",
						"load_balancer_port": "443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "8080",
						"load_balancer_port": "8080",
					}),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "subnets.*", "numspot_subnet.test_new", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "security_groups.*", "numspot_security_group.test_new", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.test", "backend_vm_ids.*", "numspot_vm.test_new", "id"),
					resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.NotEqual(t, resourceId, v) {
							return fmt.Errorf("Id should have changed")
						}
						resourceId = v
						return nil
					}),
				),
			},
			// <== If resource has optional dependencies ==>
			{ // 7 - Reset the resource to initial state (resource tied to a subresource) in prevision of next test
				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "test" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_vm" "test" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-2c4r"

  subnet_id = numspot_subnet.test.id
  tags = [
    {
      key   = "name"
      value = "terraform-loadbalancer-acctest"
    }
  ]
}

resource "numspot_load_balancer" "test" {
  name = "elb-terraform-test"
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    }
  ]

  subnets         = [numspot_subnet.test.id]
  security_groups = [numspot_security_group.test.id]
  backend_vm_ids  = [numspot_vm.test.id]

  type = "internal"

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume"
    }
  ]
}`,
			},
			{ // 8 - Update testing With delete of dependency resource and without Replacing the resource
				// This test is useful to check wether or not the deletion of the dependencies and then the update of the main resource works properly (empty dependency)

				Config: `
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "test" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}
resource "numspot_load_balancer" "test" {
  name    = "elb-terraform-test"
  subnets = [numspot_subnet.test.id]

  listeners = [
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 8080
      load_balancer_port     = 8080
      load_balancer_protocol = "TCP"
    }
  ]

  type = "internal"

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-Volume-Updated"
    }
  ]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "tags.*", map[string]string{
						"key":   "name",
						"value": "Terraform-Test-Volume-Updated",
					}),
					resource.TestCheckResourceAttr("numspot_load_balancer.test", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "443",
						"load_balancer_port": "443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.test", "listeners.*", map[string]string{
						"backend_port":       "8080",
						"load_balancer_port": "8080",
					}),
					resource.TestCheckResourceAttrWith("numspot_load_balancer.test", "id", func(v string) error {
						if !assert.NotEmpty(t, v) {
							return fmt.Errorf("Id field should not be empty")
						}
						if !assert.Equal(t, resourceId, v) {
							return fmt.Errorf("Id should be unchanged")
						}
						resourceId = v
						return nil
					}),
				),
			},
		},
	})
}
