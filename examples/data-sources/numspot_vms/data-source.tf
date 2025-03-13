resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "test" {
  image_id  = "ami-0b7df82c"
  vm_type   = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id

}

data "numspot_vms" "testdata" {
  ids = [numspot_vm.test.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_vms.testdata.items.0.id"
  }
}