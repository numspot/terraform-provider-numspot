---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_image Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_image (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `architecture` (String) The architecture of the OMI (by default, `i386` if you specified the `FileLocation` or `RootDeviceName` parameter).
- `block_device_mappings` (Attributes List) One or more block device mappings. (see [below for nested schema](#nestedatt--block_device_mappings))
- `description` (String) A description for the new OMI.
- `image_name` (String) A unique name for the new OMI.<br />
Constraints: 3-128 alphanumeric characters, underscores (_), spaces ( ), parentheses (()), slashes (/), periods (.), or dashes (-).
- `no_reboot` (Boolean) If false, the VM shuts down before creating the OMI and then reboots. If true, the VM does not.
- `product_codes` (List of String) The product codes associated with the OMI.
- `root_device_name` (String) The name of the root device. You must specify only one of the following parameters: `FileLocation`, `RootDeviceName`, `SourceImageId` or `VmId`.
- `source_image_id` (String) The ID of the OMI you want to copy. You must specify only one of the following parameters: `FileLocation`, `RootDeviceName`, `SourceImageId` or `VmId`.
- `source_region_name` (String) The name of the source Region, which must be the same as the Region of your account.
- `vm_id` (String) The ID of the VM from which you want to create the OMI. You must specify only one of the following parameters: `FileLocation`, `RootDeviceName`, `SourceImageId` or `VmId`.

### Read-Only

- `creation_date` (String) The date and time of creation of the OMI, in ISO 8601 date-time format.
- `id` (String) The ID of the OMI.
- `name` (String) The name of the OMI.
- `root_device_type` (String) The type of root device used by the OMI (always `bsu`).
- `state` (String) The state of the OMI (`pending` \| `available` \| `failed`).
- `state_comment` (Attributes) Information about the change of state. (see [below for nested schema](#nestedatt--state_comment))
- `type` (String) The type of the OMI.

<a id="nestedatt--block_device_mappings"></a>
### Nested Schema for `block_device_mappings`

Optional:

- `bsu` (Attributes) Information about the BSU volume to create. (see [below for nested schema](#nestedatt--block_device_mappings--bsu))
- `device_name` (String) The device name for the volume. For a root device, you must use `/dev/sda1`. For other volumes, you must use `/dev/sdX`, `/dev/sdXX`, `/dev/xvdX`, or `/dev/xvdXX` (where the first `X` is a letter between `b` and `z`, and the second `X` is a letter between `a` and `z`).
- `virtual_device_name` (String) The name of the virtual device (`ephemeralN`).

<a id="nestedatt--block_device_mappings--bsu"></a>
### Nested Schema for `block_device_mappings.bsu`

Optional:

- `delete_on_vm_deletion` (Boolean) By default or if set to true, the volume is deleted when terminating the VM. If false, the volume is not deleted when terminating the VM.
- `iops` (Number) The number of I/O operations per second (IOPS). This parameter must be specified only if you create an `io1` volume. The maximum number of IOPS allowed for `io1` volumes is `13000` with a maximum performance ratio of 300 IOPS per gibibyte.
- `snapshot_id` (String) The ID of the snapshot used to create the volume.
- `volume_size` (Number) The size of the volume, in gibibytes (GiB).<br />
If you specify a snapshot ID, the volume size must be at least equal to the snapshot size.<br />
If you specify a snapshot ID but no volume size, the volume is created with a size similar to the snapshot one.
- `volume_type` (String) The type of the volume (`standard` \| `io1` \| `gp2`). If not specified in the request, a `standard` volume is created.<br />
 For more information about volume types, see [About Volumes > Volume Types and IOPS](https://docs.outscale.com/en/userguide/About-Volumes.html#_volume_types_and_iops).



<a id="nestedatt--state_comment"></a>
### Nested Schema for `state_comment`

Read-Only:

- `state_code` (String) The code of the change of state.
- `state_message` (String) A message explaining the change of state.