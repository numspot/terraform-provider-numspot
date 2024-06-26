---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_nics Data Source - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_nics (Data Source)



## Example Usage

```terraform
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}


resource "numspot_nic" "test" {
  subnet_id = numspot_subnet.subnet.id
}

data "numspot_nics" "testdata" {
  ids        = [numspot_nic.test.id]
  depends_on = [numspot_nic.test]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_nics.testdata.items.0.id"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `availability_zone_names` (List of String) The Subregion in which the NIC is located.
- `descriptions` (List of String) The description of the NIC.
- `ids` (List of String) ID for ReadNics
- `is_source_dest_checked` (Boolean) (Net only) If true, the source/destination check is enabled. If false, it is disabled. This value must be false for a NAT VM to perform network address translation (NAT) in a Net.
- `link_nic_delete_on_vm_deletion` (Boolean) If true, the NIC is deleted when the VM is terminated.
- `link_nic_device_numbers` (List of Number) The device index for the NIC attachment (between `1` and `7`, both included).
- `link_nic_link_nic_ids` (List of String) The ID of the NIC to attach.
- `link_nic_states` (List of String) The state of the attachment (`attaching` \| `attached` \| `detaching` \| `detached`).
- `link_nic_vm_ids` (List of String) The ID of the VM.
- `link_public_ip_ids` (List of String) (Required in a Net) The ID representing the association of the public IP with the VM or the NIC.
- `link_public_ip_public_ip_ids` (List of String) The allocation ID of the public IP.
- `link_public_ip_public_ips` (List of String) The public IP associated with the NIC.
- `mac_addresses` (List of String) The Media Access Control (MAC) address of the NIC.
- `private_dns_names` (List of String) The name of the private DNS.
- `private_ips_is_primary` (Boolean) If true, the IP is the primary private IP of the NIC.
- `private_ips_link_public_ip_public_ips` (List of String) The public IPs associated with the private IPs.
- `private_ips_private_ips` (List of String) The private IP of the NIC.
- `security_group_ids` (List of String) The ID of the security group.
- `security_group_names` (List of String) The name of the security group.
- `states` (List of String) The state of the NIC (`available` \| `attaching` \| `in-use` \| `detaching`).
- `subnet_ids` (List of String) The ID of the Subnet.
- `tag_keys` (List of String) The key of the tag, with a minimum of 1 character.
- `tag_values` (List of String) The value of the tag, between 0 and 255 characters.
- `tags` (List of String) The key/value combination of the tags associated with the DHCP options sets, in the following format: "Filters":{"Tags":["TAGKEY=TAGVALUE"]}.
- `vpc_ids` (List of String) The ID of the Net for the NIC.

### Read-Only

- `items` (Attributes List) (see [below for nested schema](#nestedatt--items))

<a id="nestedatt--items"></a>
### Nested Schema for `items`

Required:

- `subnet_id` (String) The ID of the Subnet in which you want to create the NIC.

Optional:

- `description` (String) A description for the NIC.
- `private_ips` (Attributes List) The primary private IP for the NIC.<br />
This IP must be within the IP range of the Subnet that you specify with the `SubnetId` attribute.<br />
If you do not specify this attribute, a random private IP is selected within the IP range of the Subnet. (see [below for nested schema](#nestedatt--items--private_ips))
- `tags` (Attributes List) One or more tags associated with the resource. (see [below for nested schema](#nestedatt--items--tags))

Read-Only:

- `availability_zone_name` (String) The Subregion in which the NIC is located.
- `id` (String) The ID of the NIC.
- `is_source_dest_checked` (Boolean) (Net only) If true, the source/destination check is enabled. If false, it is disabled. This value must be false for a NAT VM to perform network address translation (NAT) in a Net.
- `link_nic` (Attributes) Information about the NIC attachment. (see [below for nested schema](#nestedatt--items--link_nic))
- `link_public_ip` (Attributes) Information about the public IP association. (see [below for nested schema](#nestedatt--items--link_public_ip))
- `mac_address` (String) The Media Access Control (MAC) address of the NIC.
- `private_dns_name` (String) The name of the private DNS.
- `security_groups` (Attributes List) One or more IDs of security groups for the NIC. (see [below for nested schema](#nestedatt--items--security_groups))
- `state` (String) The state of the NIC (`available` \| `attaching` \| `in-use` \| `detaching`).
- `vpc_id` (String) The ID of the Net for the NIC.

<a id="nestedatt--items--private_ips"></a>
### Nested Schema for `items.private_ips`

Optional:

- `is_primary` (Boolean) If true, the IP is the primary private IP of the NIC.
- `private_ip` (String) The private IP of the NIC.

Read-Only:

- `link_public_ip` (Attributes) Information about the public IP association. (see [below for nested schema](#nestedatt--items--private_ips--link_public_ip))
- `private_dns_name` (String) The name of the private DNS.

<a id="nestedatt--items--private_ips--link_public_ip"></a>
### Nested Schema for `items.private_ips.link_public_ip`

Read-Only:

- `id` (String) (Required in a Net) The ID representing the association of the public IP with the VM or the NIC.
- `public_dns_name` (String) The name of the public DNS.
- `public_ip` (String) The public IP associated with the NIC.
- `public_ip_id` (String) The allocation ID of the public IP.



<a id="nestedatt--items--tags"></a>
### Nested Schema for `items.tags`

Required:

- `key` (String) The key of the tag, with a minimum of 1 character.
- `value` (String) The value of the tag, between 0 and 255 characters.


<a id="nestedatt--items--link_nic"></a>
### Nested Schema for `items.link_nic`

Read-Only:

- `delete_on_vm_deletion` (Boolean) If true, the NIC is deleted when the VM is terminated.
- `device_number` (Number) The device index for the NIC attachment (between `1` and `7`, both included).
- `id` (String) The ID of the NIC to attach.
- `state` (String) The state of the attachment (`attaching` \| `attached` \| `detaching` \| `detached`).
- `vm_id` (String) The ID of the VM.


<a id="nestedatt--items--link_public_ip"></a>
### Nested Schema for `items.link_public_ip`

Read-Only:

- `id` (String) (Required in a Net) The ID representing the association of the public IP with the VM or the NIC.
- `public_dns_name` (String) The name of the public DNS.
- `public_ip` (String) The public IP associated with the NIC.
- `public_ip_id` (String) The allocation ID of the public IP.


<a id="nestedatt--items--security_groups"></a>
### Nested Schema for `items.security_groups`

Read-Only:

- `security_group_id` (String) The ID of the security group.
- `security_group_name` (String) The name of the security group.
