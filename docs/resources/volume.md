---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_volume Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_volume (Resource)



## Example Usage

```terraform
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "vm" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
  tags = [
    {
      key   = "name"
      value = "terraform-volume-acctest"
    }
  ]
}

resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "cloudgouv-eu-west-1a"
  tags = [
    {
      key   = "name"
      value = "My Volume"
    }
  ]
  link_vm = {
    vm_id       = numspot_vm.vm.id
    device_name = "/dev/sdb"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `availability_zone_name` (String) The Subregion in which you want to create the volume.

### Optional

- `iops` (Number) The number of I/O operations per second (IOPS). This parameter must be specified only if you create an `io1` volume. The maximum number of IOPS allowed for `io1` volumes is `13000` with a maximum performance ratio of 300 IOPS per gibibyte.
- `link_vm` (Attributes) VM the Volume will be linked to. To unlink a Volume from a VM, the VM will need to be restarded. (see [below for nested schema](#nestedatt--link_vm))
- `replace_volume_on_downsize` (Boolean) If replace_volume_on_downsize is set to 'true' and volume size is reduced, the volume will be deleted and recreated.  WARNING : All data on the volume will be lost. Default is false
- `size` (Number) The size of the volume, in gibibytes (GiB). The maximum allowed size for a volume is 14901 GiB. This parameter is required if the volume is not created from a snapshot (`SnapshotId` unspecified).
- `snapshot_id` (String) The ID of the snapshot from which you want to create the volume.
- `tags` (Attributes List) One or more tags associated with the resource. (see [below for nested schema](#nestedatt--tags))
- `type` (String) The type of volume you want to create (`io1` \| `gp2` \ | `standard`). If not specified, a `standard` volume is created.<br />

### Read-Only

- `creation_date` (String) The date and time of creation of the volume.
- `id` (String) The ID of the volume.
- `linked_volumes` (Attributes List) Information about your volume attachment. (see [below for nested schema](#nestedatt--linked_volumes))
- `state` (String) The state of the volume (`creating` \| `available` \| `in-use` \| `updating` \| `deleting` \| `error`).

<a id="nestedatt--link_vm"></a>
### Nested Schema for `link_vm`

Optional:

- `device_name` (String) The name of the device. For a root device, you must use /dev/sda1. For other volumes, you must use /dev/sdX, /dev/sdXX, /dev/xvdX, or /dev/xvdXX (where the first X is a letter between b and z, and the second X is a letter between a and z).
- `vm_id` (String) The ID of the VM you want to attach the volume to.


<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Required:

- `key` (String) The key of the tag, with a minimum of 1 character.
- `value` (String) The value of the tag, between 0 and 255 characters.


<a id="nestedatt--linked_volumes"></a>
### Nested Schema for `linked_volumes`

Read-Only:

- `delete_on_vm_deletion` (Boolean) If true, the volume is deleted when terminating the VM. If false, the volume is not deleted when terminating the VM.
- `device_name` (String) The name of the device.
- `id` (String) The ID of the volume.
- `state` (String) The state of the attachment of the volume (`attaching` \| `detaching` \| `attached` \| `detached`).
- `vm_id` (String) The ID of the VM.