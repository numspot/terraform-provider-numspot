package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/acctest"
)

func TestAccLoadBalancerResource(t *testing.T) {
	acct := acctest.NewAccTest(t, false, "")
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	loadBalancerDependencies := `
resource "numspot_vpc" "terraform-dep-vpc-lb" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}

resource "numspot_subnet" "terraform-dep-subnet-lb" {
  vpc_id   = numspot_vpc.terraform-dep-vpc-lb.id
  ip_range = "10.101.1.0/24"
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}

resource "numspot_vm" "terraform-dep-vm-lb" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-2c4r"

  subnet_id = numspot_subnet.terraform-dep-subnet-lb.id
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}
`

	loadBalancerUpdateDependencies := `
resource "numspot_vpc" "terraform-dep-vpc-lb" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}

resource "numspot_subnet" "terraform-dep-subnet-lb" {
  vpc_id   = numspot_vpc.terraform-dep-vpc-lb.id
  ip_range = "10.101.1.0/24"
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}

resource "numspot_security_group" "terraform-dep-sg-lb" {
  vpc_id      = numspot_vpc.terraform-dep-vpc-lb.id
  name        = "terraform acctest lb name"
  description = "terraform acctest lb description"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}

resource "numspot_vm" "terraform-dep-vm-lb" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-2c4r"

  subnet_id = numspot_subnet.terraform-dep-subnet-lb.id
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}
`

	loadBalancerReplaceDependencies := `
resource "numspot_vpc" "terraform-dep-vpc-lb" {
  ip_range = "10.101.0.0/16"
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}

resource "numspot_internet_gateway" "terraform-dep-igw-lb" {
  vpc_id = numspot_vpc.terraform-dep-vpc-lb.id

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest"
  }]
}

resource "numspot_subnet" "terraform-dep-subnet-replace-lb" {
  vpc_id   = numspot_vpc.terraform-dep-vpc-lb.id
  ip_range = "10.101.1.0/24"
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}

resource "numspot_security_group" "terraform-dep-sg-lb" {
  vpc_id      = numspot_vpc.terraform-dep-vpc-lb.id
  name        = "terraform acctest lb name"
  description = "terraform acctest lb description"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}

resource "numspot_vm" "terraform-dep-vm-lb" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-2c4r"

  subnet_id = numspot_subnet.terraform-dep-subnet-replace-lb.id
  tags = [
    {
      key   = "name"
      value = "terraform-load-balancer-acctest"
    }
  ]
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			// Step 1 - Create simple load-balancer
			{
				Config: loadBalancerDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest" {
  name = "elb-terraform-test"
  type = "internal"

  subnets = [numspot_subnet.terraform-dep-subnet-lb.id]

  listeners = [{
    backend_port           = 80
    load_balancer_port     = 80
    load_balancer_protocol = "TCP"
  }]

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "type", "internal"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "subnets.*", "numspot_subnet.terraform-dep-subnet-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "listeners.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest",
					}),
				),
			},
			// Step 2 - Import
			{
				ResourceName:            "numspot_load_balancer.terraform-lb-acctest",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// Step 3 - Link load-balancer Backend
			{
				Config: loadBalancerDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest" {
  name = "elb-terraform-test"
  type = "internal"

  subnets        = [numspot_subnet.terraform-dep-subnet-lb.id]
  backend_vm_ids = [numspot_vm.terraform-dep-vm-lb.id]

  listeners = [{
    backend_port           = 80
    load_balancer_port     = 80
    load_balancer_protocol = "TCP"
  }]

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "type", "internal"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "subnets.*", "numspot_subnet.terraform-dep-subnet-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "backend_vm_ids.*", "numspot_vm.terraform-dep-vm-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "listeners.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest",
					}),
				),
			},
			// Step 4 - Link load-balancer HealthCheck
			{
				Config: loadBalancerDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest" {
  name = "elb-terraform-test"
  type = "internal"

  subnets        = [numspot_subnet.terraform-dep-subnet-lb.id]
  backend_vm_ids = [numspot_vm.terraform-dep-vm-lb.id]

  listeners = [{
    backend_port           = 80
    load_balancer_port     = 80
    load_balancer_protocol = "TCP"
  }]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/"
    port                = 80
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "type", "internal"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "subnets.*", "numspot_subnet.terraform-dep-subnet-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "backend_vm_ids.*", "numspot_vm.terraform-dep-vm-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "listeners.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.healthy_threshold", "10"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.check_interval", "30"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.path", "/"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.port", "80"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.protocol", "HTTP"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.timeout", "5"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.unhealthy_threshold", "5"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest",
					}),
				),
			},
			// Step 5 - Update load-balancer attributes
			{
				Config: loadBalancerUpdateDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest" {
  name = "elb-terraform-test"
  type = "internal"

  subnets         = [numspot_subnet.terraform-dep-subnet-lb.id]
  backend_vm_ids  = [numspot_vm.terraform-dep-vm-lb.id]
  security_groups = [numspot_security_group.terraform-dep-sg-lb.id]

  listeners = [{
    backend_port           = 80
    load_balancer_port     = 80
    load_balancer_protocol = "TCP"
  }]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/"
    port                = 80
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest-update"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "type", "internal"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "subnets.*", "numspot_subnet.terraform-dep-subnet-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "backend_vm_ids.*", "numspot_vm.terraform-dep-vm-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "security_groups.*", "numspot_security_group.terraform-dep-sg-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "listeners.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.healthy_threshold", "10"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.check_interval", "30"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.path", "/"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.port", "80"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.protocol", "HTTP"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.timeout", "5"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.unhealthy_threshold", "5"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest-update",
					}),
				),
			},
			// Step 6 - Replace load-balancer attributes
			{
				Config: loadBalancerReplaceDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest" {
  name = "elb-terraform-test"
  type = "internet-facing"

  subnets         = [numspot_subnet.terraform-dep-subnet-replace-lb.id]
  backend_vm_ids  = [numspot_vm.terraform-dep-vm-lb.id]
  security_groups = [numspot_security_group.terraform-dep-sg-lb.id]

  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    }
  ]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/"
    port                = 80
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest-replace"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "type", "internet-facing"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "subnets.*", "numspot_subnet.terraform-dep-subnet-replace-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "backend_vm_ids.*", "numspot_vm.terraform-dep-vm-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "security_groups.*", "numspot_security_group.terraform-dep-sg-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "443",
						"load_balancer_port":     "443",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.healthy_threshold", "10"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.check_interval", "30"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.path", "/"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.port", "80"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.protocol", "HTTP"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.timeout", "5"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.unhealthy_threshold", "5"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest-replace",
					}),
				),
			},
			// Step 7 - Reset
			{
				Config: ` `,
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Step 8 - Create load-balancer with attributes
			{
				Config: loadBalancerReplaceDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest" {
  name = "elb-terraform-test"
  type = "internet-facing"

  subnets         = [numspot_subnet.terraform-dep-subnet-replace-lb.id]
  backend_vm_ids  = [numspot_vm.terraform-dep-vm-lb.id]
  security_groups = [numspot_security_group.terraform-dep-sg-lb.id]

  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    }
  ]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/"
    port                = 80
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest-replace"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "type", "internet-facing"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "subnets.*", "numspot_subnet.terraform-dep-subnet-replace-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "backend_vm_ids.*", "numspot_vm.terraform-dep-vm-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "security_groups.*", "numspot_security_group.terraform-dep-sg-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "443",
						"load_balancer_port":     "443",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.healthy_threshold", "10"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.check_interval", "30"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.path", "/"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.port", "80"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.protocol", "HTTP"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.timeout", "5"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.unhealthy_threshold", "5"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest-replace",
					}),
				),
			},
			// Step 9 - Unlink load-balancer Backend
			{
				Config: loadBalancerReplaceDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest" {
  name = "elb-terraform-test"
  type = "internet-facing"

  subnets         = [numspot_subnet.terraform-dep-subnet-replace-lb.id]
  security_groups = [numspot_security_group.terraform-dep-sg-lb.id]

  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    }
  ]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/"
    port                = 80
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest-replace"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "type", "internet-facing"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "subnets.*", "numspot_subnet.terraform-dep-subnet-replace-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "security_groups.*", "numspot_security_group.terraform-dep-sg-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "443",
						"load_balancer_port":     "443",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.healthy_threshold", "10"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.check_interval", "30"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.path", "/"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.port", "80"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.protocol", "HTTP"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.timeout", "5"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "health_check.unhealthy_threshold", "5"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest-replace",
					}),
				),
			},
			// Step 10 - Unlink load-balancer HealthChecks
			{
				Config: loadBalancerReplaceDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest" {
  name = "elb-terraform-test"
  type = "internet-facing"

  subnets         = [numspot_subnet.terraform-dep-subnet-replace-lb.id]
  security_groups = [numspot_security_group.terraform-dep-sg-lb.id]

  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    }
  ]

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest-replace"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "type", "internet-facing"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "subnets.*", "numspot_subnet.terraform-dep-subnet-replace-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest", "security_groups.*", "numspot_security_group.terraform-dep-sg-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "listeners.*", map[string]string{
						"backend_port":           "443",
						"load_balancer_port":     "443",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest-replace",
					}),
				),
			},
			// Step 11 - Recreate load-balancer
			{
				Config: loadBalancerReplaceDependencies + `
resource "numspot_load_balancer" "terraform-lb-acctest-recreate" {
  name = "elb-terraform-test"
  type = "internet-facing"

  subnets         = [numspot_subnet.terraform-dep-subnet-replace-lb.id]
  backend_vm_ids  = [numspot_vm.terraform-dep-vm-lb.id]
  security_groups = [numspot_security_group.terraform-dep-sg-lb.id]

  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    },
    {
      backend_port           = 443
      load_balancer_port     = 443
      load_balancer_protocol = "TCP"
    }
  ]

  health_check = {
    healthy_threshold   = 10
    check_interval      = 30
    path                = "/"
    port                = 80
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [{
    key   = "name"
    value = "terraform-load-balancer-acctest-recreate"
  }]
}`,

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "name", "elb-terraform-test"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "type", "internet-facing"),

					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest-recreate", "subnets.*", "numspot_subnet.terraform-dep-subnet-replace-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest-recreate", "backend_vm_ids.*", "numspot_vm.terraform-dep-vm-lb", "id"),
					resource.TestCheckTypeSetElemAttrPair("numspot_load_balancer.terraform-lb-acctest-recreate", "security_groups.*", "numspot_security_group.terraform-dep-sg-lb", "id"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "listeners.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest-recreate", "listeners.*", map[string]string{
						"backend_port":           "80",
						"load_balancer_port":     "80",
						"load_balancer_protocol": "TCP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest-recreate", "listeners.*", map[string]string{
						"backend_port":           "443",
						"load_balancer_port":     "443",
						"load_balancer_protocol": "TCP",
					}),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "health_check.healthy_threshold", "10"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "health_check.check_interval", "30"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "health_check.path", "/"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "health_check.port", "80"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "health_check.protocol", "HTTP"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "health_check.timeout", "5"),
					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "health_check.unhealthy_threshold", "5"),

					resource.TestCheckResourceAttr("numspot_load_balancer.terraform-lb-acctest-recreate", "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("numspot_load_balancer.terraform-lb-acctest-recreate", "tags.*", map[string]string{
						"key":   "name",
						"value": "terraform-load-balancer-acctest-recreate",
					}),
				),
			},
		},
	})
}
