## Internal Load Balancer

resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "security-group" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "My SG Group"
  description = "terraform-vm-tests-sg-description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_protocol     = "tcp"
    }
  ]
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_protocol     = "tcp"
    },
  ]
}

resource "numspot_vm" "vm" {
  image_id  = "ami-42e5719f"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_load_balancer" "load-balancer" {
  name = "load-balancer"
  type = "internal"

  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      backend_protocol       = "TCP"
      load_balancer_protocol = "TCP"
    }
  ]

  subnets         = [numspot_subnet.subnet.id]
  security_groups = [numspot_security_group.security-group.id]
  backend_vm_ids  = [numspot_vm.vm.id]
  backend_ips     = ["192.0.2.0"]


  health_check = {
    check_interval      = 30
    healthy_threshold   = 10
    path                = "/index.html"
    port                = 8080
    protocol            = "HTTPS"
    timeout             = 5
    unhealthy_threshold = 5
  }

  tags = [
    {
      key   = "Env"
      value = "Prod"
    }
  ]
}

## Public Load Balancer

resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "internet-gateway" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_subnet" "subnet" {
  vpc_id                  = numspot_vpc.vpc.id
  ip_range                = "10.101.1.0/24"
  map_public_ip_on_launch = true
}

resource "numspot_security_group" "security-group" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_protocol     = "tcp"
    }
  ]
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_route_table" "route_table" {
  vpc_id    = numspot_vpc.vpc.id
  subnet_id = numspot_subnet.subnet.id

  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.internet-gateway.id
    }
  ]
}

resource "numspot_vm" "vm" {
  image_id = "ami-8ef5b47e"
  type     = "ns-cus6-4c8r"

  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_load_balancer" "load-balancer" {
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

  subnets         = [numspot_subnet.subnet.id]
  security_groups = [numspot_security_group.security-group.id]
  backend_vm_ids  = [numspot_vm.vm.id]

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

  depends_on = [numspot_internet_gateway.internet-gateway]
}
