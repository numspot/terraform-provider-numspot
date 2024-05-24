
resource "numspot_vm" "test" {
  image_id = "ami-026ce760"
  vm_type  = "ns-cus6-2c4r"
}

data "numspot_vms" "testdata" {
  ids        = [numspot_vm.test.id]
  depends_on = [numspot_vm.test]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_vms.testdata.items.0.id"
  }
}