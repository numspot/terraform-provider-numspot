resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_security_group" "sg" {
  net_id      = numspot_vpc.vpc.id
  name        = "terraform-vm-tests-sg-name"
  description = "terraform-vm-tests-sg-description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

resource "numspot_internet_gateway" "igw" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_route_table" "rt" {
  vpc_id    = numspot_vpc.vpc.id
  subnet_id = numspot_subnet.subnet.id

  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.igw.id
    }
  ]
}

resource "numspot_vm" "test" {
  image_id = "ami-0b7df82c"
  vm_type  = "tinav6.c1r1p3"

  subnet_id          = numspot_subnet.subnet.id
  security_group_ids = [numspot_security_group.sg.id]

  depends_on = [numspot_security_group.sg]
}

resource "numspot_load_balancer" "testlb" {
  name = "http-load-balancer"
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"
    }
  ]

  subnets = [numspot_subnet.subnet.id]

  type = "internal"

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
      key   = "env"
      value = "prod"
    }
  ]

  backend_vm_ids = [numspot_vm.test.id]
}