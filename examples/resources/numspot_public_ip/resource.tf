resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "vm" {
  image_id  = "ami-12345678"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_internet_gateway" "ig" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_public_ip" "public_ip" {
  vm_id = numspot_vm.vm.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
  depends_on = [numspot_internet_gateway.ig]
}