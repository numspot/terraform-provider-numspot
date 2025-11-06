# Image from image
resource "numspot_image" "image_from_image" {
  name               = "image-from-image"
  source_image_id    = "ami-foobar"
  source_region_name = "eu-west-2"
  access = {
    is_public = "false"
  }
  tags = [
    {
      key   = "Name"
      value = "My Image"
    }
  ]
}

# Image from VM
data "numspot_vm" "vm" {
  # ...
}

resource "numspot_image" "image_from_vm" {
  name  = "image-from-vm"
  vm_id = numspot_vm.vm.id
}

# Image from Snapshot
data "numspot_snapshot" "snapshot" {
  # ...
}

resource "numspot_image" "image_from_snapshot" {
  name             = "image-from-updated"
  root_device_name = "/dev/sda1"
  block_device_mappings = [
    {
      device_name = "/dev/sda1"
      bsu = {
        snapshot_id           = numspot_snapshot.snapshot.id
        volume_size           = 120
        volume_type           = "io1"
        iops                  = 150
        delete_on_vm_deletion = true
      }
    }
  ]
}
