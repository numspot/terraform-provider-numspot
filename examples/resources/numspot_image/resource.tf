# Image from image
resource "numspot_image" "image" {
  name               = "image-from-image"
  source_image_id    = "ami-foobar"
  source_region_name = "cloudgouv-eu-west-1"
}

# Image from VM
resource "numspot_vm" "vm" {
  image_id = numspot_image.image.id
  vm_type  = "tinav6.c1r1p3"
}

resource "numspot_image" "image_from_vm" {
  name  = "image-from-vm-imahe"
  vm_id = numspot_vm.vm.id
}