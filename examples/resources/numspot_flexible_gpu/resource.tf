resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_vm" "test" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_flexible_gpu" "gpu" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "eu-west-2a"
  vm_id                  = numspot_vm.test.id
}
