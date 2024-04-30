resource "numspot_vm" "vm" {
  image_id = "ami-060e019f"
  vm_type  = "tinav6.c1r1p3"
}

resource "numspot_public_ip" "test" {
  vm_id = numspot_vm.vm.id
}