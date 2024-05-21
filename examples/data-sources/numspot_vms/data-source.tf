
resource "numspot_vm" "test" {
  image_id = "ami-026ce760"
  vm_type  = "ns-cus6-2c4r"
}

data "numspot_vms" "testdata" {
  ids        = [numspot_vm.test.id]
  depends_on = [numspot_vm.test]
}

