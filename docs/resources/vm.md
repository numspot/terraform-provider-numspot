---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_vm Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_vm (Resource)



## Example Usage

```terraform
resource "numspot_vm" "example" {
  image_id = "ami-12345678"
}

resource "numspot_vm" "example" {
  image_id           = "ami-12345678"
  vm_type            = "tinav5.c1r1p2"
  keypair_name       = "keypair-example"
  security_group_ids = ["sg-12345678"]
  user_data          = "..."
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `image_id` (String) The ID of the OMI used to create the VM. You can find the list of OMIs by calling the [ReadImages](#readimages) method.

### Optional

- `block_device_mappings` (Attributes List) One or more block device mappings. (see [below for nested schema](#nestedatt--block_device_mappings))
- `boot_on_creation` (Boolean) By default or if true, the VM is started on creation. If false, the VM is stopped on creation.
- `bsu_optimized` (Boolean) This parameter is not available. It is present in our API for the sake of historical compatibility with AWS.
- `client_token` (String) A unique identifier which enables you to manage the idempotency.
- `deletion_protection` (Boolean) If true, you cannot delete the VM unless you change this parameter back to false.
- `keypair_name` (String) The name of the keypair.
- `nested_virtualization` (Boolean) (dedicated tenancy only) If true, nested virtualization is enabled. If false, it is disabled.
- `nics` (Attributes List) One or more NICs. If you specify this parameter, you must not specify the `SubnetId` and `SubregionName` parameters. You also must define one NIC as the primary network interface of the VM with `0` as its device number. (see [below for nested schema](#nestedatt--nics))
- `performance` (String) The performance of the VM (`medium` \| `high` \|  `highest`). By default, `high`. This parameter is ignored if you specify a performance flag directly in the `VmType` parameter.
- `placement` (Attributes) Information about the placement of the VM. (see [below for nested schema](#nestedatt--placement))
- `private_ips` (List of String) One or more private IPs of the VM.
- `security_group_ids` (List of String) One or more IDs of security group for the VMs.
- `security_groups` (List of String) One or more names of security groups for the VMs.
- `subnet_id` (String) The ID of the Subnet in which you want to create the VM. If you specify this parameter, you must not specify the `Nics` parameter.
- `user_data` (String) Data or script used to add a specific configuration to the VM. It must be Base64-encoded and is limited to 500 kibibytes (KiB).
- `vm_initiated_shutdown_behavior` (String) The VM behavior when you stop it. By default or if set to `stop`, the VM stops. If set to `restart`, the VM stops then automatically restarts. If set to `terminate`, the VM stops and is terminated.
- `vm_type` (String) The type of VM. You can specify a TINA type (in the `tinavW.cXrYpZ` or `tinavW.cXrY` format), or an AWS type (for example, `t2.small`, which is the default value).<br />
If you specify an AWS type, it is converted in the background to its corresponding TINA type, but the AWS type is still returned. If the specified or converted TINA type includes a performance flag, this performance flag is applied regardless of the value you may have provided in the `Performance` parameter. For more information, see [Instance Types](https://docs.outscale.com/en/userguide/Instance-Types.html).
- `vms_count` (Number) The minimum number of VMs you want to create. If this number of VMs cannot be created, no VMs are created.

### Read-Only

- `architecture` (String) The architecture of the VM (`i386` \| `x86_64`).
- `creation_date` (String) The date and time of creation of the VM.
- `hypervisor` (String) The hypervisor type of the VMs (`ovm` \| `xen`).
- `id` (String) The ID of the VM.
- `initiated_shutdown_behavior` (String) The VM behavior when you stop it. If set to `stop`, the VM stops. If set to `restart`, the VM stops then automatically restarts. If set to `terminate`, the VM stops and is deleted.
- `is_source_dest_checked` (Boolean) (Net only) If true, the source/destination check is enabled. If false, it is disabled. This value must be false for a NAT VM to perform network address translation (NAT) in a Net.
- `launch_number` (Number) The number for the VM when launching a group of several VMs (for example, `0`, `1`, `2`, and so on).
- `net_id` (String) The ID of the Net in which the VM is running.
- `os_family` (String) Indicates the operating system (OS) of the VM.
- `private_dns_name` (String) The name of the private DNS.
- `private_ip` (String) The primary private IP of the VM.
- `product_codes` (List of String) The product codes associated with the OMI used to create the VM.
- `public_dns_name` (String) The name of the public DNS.
- `public_ip` (String) The public IP of the VM.
- `reservation_id` (String) The reservation ID of the VM.
- `root_device_name` (String) The name of the root device for the VM (for example, `/dev/vda1`).
- `root_device_type` (String) The type of root device used by the VM (always `bsu`).
- `state` (String) The state of the VM (`pending` \| `running` \| `stopping` \| `stopped` \| `shutting-down` \| `terminated` \| `quarantine`).
- `state_reason` (String) The reason explaining the current state of the VM.
- `type` (String) The type of VM. For more information, see [Instance Types](https://docs.outscale.com/en/userguide/Instance-Types.html).

<a id="nestedatt--block_device_mappings"></a>
### Nested Schema for `block_device_mappings`

Optional:

- `bsu` (Attributes) Information about the BSU volume to create. (see [below for nested schema](#nestedatt--block_device_mappings--bsu))
- `device_name` (String) The device name for the volume. For a root device, you must use `/dev/sda1`. For other volumes, you must use `/dev/sdX`, `/dev/sdXX`, `/dev/xvdX`, or `/dev/xvdXX` (where the first `X` is a letter between `b` and `z`, and the second `X` is a letter between `a` and `z`).
- `no_device` (String) Removes the device which is included in the block device mapping of the OMI.
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

Read-Only:

- `link_date` (String) The date and time of attachment of the volume to the VM, in ISO 8601 date-time format.
- `state` (String) The state of the volume.
- `volume_id` (String) The ID of the volume.



<a id="nestedatt--nics"></a>
### Nested Schema for `nics`

Optional:

- `delete_on_vm_deletion` (Boolean) If true, the NIC is deleted when the VM is terminated. You can specify this parameter only for a new NIC. To modify this value for an existing NIC, see [UpdateNic](#updatenic).
- `description` (String) The description of the NIC, if you are creating a NIC when creating the VM.
- `device_number` (Number) The index of the VM device for the NIC attachment (between `0` and `7`, both included). This parameter is required if you create a NIC when creating the VM.
- `nic_id` (String) The ID of the NIC, if you are attaching an existing NIC when creating a VM.
- `private_ips` (Attributes List) One or more private IPs to assign to the NIC, if you create a NIC when creating a VM. Only one private IP can be the primary private IP. (see [below for nested schema](#nestedatt--nics--private_ips))
- `secondary_private_ip_count` (Number) The number of secondary private IPs, if you create a NIC when creating a VM. This parameter cannot be specified if you specified more than one private IP in the `PrivateIps` parameter.
- `security_group_ids` (List of String) One or more IDs of security groups for the NIC, if you create a NIC when creating a VM.
- `subnet_id` (String) The ID of the Subnet for the NIC, if you create a NIC when creating a VM. This parameter is required if you create a NIC when creating the VM.

Read-Only:

- `account_id` (String) The account ID of the owner of the NIC.
- `is_source_dest_checked` (Boolean) (Net only) If true, the source/destination check is enabled. If false, it is disabled. This value must be false for a NAT VM to perform network address translation (NAT) in a Net.
- `link_nic` (Attributes) Information about the network interface card (NIC). (see [below for nested schema](#nestedatt--nics--link_nic))
- `link_public_ip` (Attributes) Information about the public IP associated with the NIC. (see [below for nested schema](#nestedatt--nics--link_public_ip))
- `mac_address` (String) The Media Access Control (MAC) address of the NIC.
- `net_id` (String) The ID of the Net for the NIC.
- `private_dns_name` (String) The name of the private DNS.
- `security_groups` (Attributes List) One or more IDs of security groups for the NIC. (see [below for nested schema](#nestedatt--nics--security_groups))
- `state` (String) The state of the NIC (`available` \| `attaching` \| `in-use` \| `detaching`).

<a id="nestedatt--nics--private_ips"></a>
### Nested Schema for `nics.private_ips`

Optional:

- `is_primary` (Boolean) If true, the IP is the primary private IP of the NIC.
- `private_ip` (String) The private IP of the NIC.

Read-Only:

- `link_public_ip` (Attributes) Information about the public IP associated with the NIC. (see [below for nested schema](#nestedatt--nics--private_ips--link_public_ip))
- `private_dns_name` (String) The name of the private DNS.

<a id="nestedatt--nics--private_ips--link_public_ip"></a>
### Nested Schema for `nics.private_ips.link_public_ip`

Read-Only:

- `public_dns_name` (String) The name of the public DNS.
- `public_ip` (String) The public IP associated with the NIC.
- `public_ip_account_id` (String) The account ID of the owner of the public IP.



<a id="nestedatt--nics--link_nic"></a>
### Nested Schema for `nics.link_nic`

Read-Only:

- `delete_on_vm_deletion` (Boolean) If true, the NIC is deleted when the VM is terminated.
- `device_number` (Number) The device index for the NIC attachment (between `1` and `7`, both included).
- `link_nic_id` (String) The ID of the NIC to attach.
- `state` (String) The state of the attachment (`attaching` \| `attached` \| `detaching` \| `detached`).


<a id="nestedatt--nics--link_public_ip"></a>
### Nested Schema for `nics.link_public_ip`

Read-Only:

- `public_dns_name` (String) The name of the public DNS.
- `public_ip` (String) The public IP associated with the NIC.
- `public_ip_account_id` (String) The account ID of the owner of the public IP.


<a id="nestedatt--nics--security_groups"></a>
### Nested Schema for `nics.security_groups`

Read-Only:

- `security_group_id` (String) The ID of the security group.
- `security_group_name` (String) The name of the security group.



<a id="nestedatt--placement"></a>
### Nested Schema for `placement`

Optional:

- `subregion_name` (String) The name of the Subregion. If you specify this parameter, you must not specify the `Nics` parameter.
- `tenancy` (String) The tenancy of the VM (`default` \| `dedicated`).