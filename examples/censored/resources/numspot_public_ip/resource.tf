resource "numspot_vm" "vm" {
  image_id = "ami-060e019f"
  vm_type  = "ns-cus6-2c4r"
}

resource "numspot_public_ip" "test" {
  vm_id = numspot_vm.vm.id
}