resource "numspot_vm" "vm" {
  image_id = "ami-0b7df82c"
  type     = "ns-mus6-2c16r"
}

resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1a"
  delete_on_vm_deletion  = true
  vm_id                  = numspot_vm.vm.id
}