resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "vm" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "My Volume"
    }
  ]
  link_vm = {
    vm_id       = numspot_vm.vm.id
    device_name = "/dev/sdb"
  }
}