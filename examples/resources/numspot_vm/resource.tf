resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_security_group" "sg" {
  vpc_id      = numspot_vpc.vpc.id
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
  vpc_id = numspot_vpc.net.id
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

resource "numspot_public_ip" "public_ip" {
  vm_id      = numspot_vm.vm.id
  depends_on = [numspot_route_table.rt]
}

resource "numspot_vm" "vm" {
  image_id                       = "ami-0987a84b"
  type                           = "ns-eco6-2c8r"
  vm_initiated_shutdown_behavior = "stop"

  placement = {
    tenancy                = "default"
    availability_zone_name = "cloudgouv-eu-west-1a"
  }

  subnet_id          = numspot_subnet.subnet.id
  security_group_ids = [numspot_security_group.sg.id]

  tags = [
    {
      key   = "name"
      value = "Terraform-Test-VM"
    }
  ]
}