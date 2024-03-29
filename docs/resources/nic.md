---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_nic Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_nic (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `subnet_id` (String) The ID of the Subnet in which you want to create the NIC.

### Optional

- `description` (String) A description for the NIC.
- `private_ips` (Attributes List) The primary private IP for the NIC.<br />
This IP must be within the IP range of the Subnet that you specify with the `SubnetId` attribute.<br />
If you do not specify this attribute, a random private IP is selected within the IP range of the Subnet. (see [below for nested schema](#nestedatt--private_ips))
- `security_group_ids` (List of String) One or more IDs of security groups for the NIC.

### Read-Only

- `account_id` (String) The account ID of the owner of the NIC.
- `id` (String) The ID of the NIC.
- `is_source_dest_checked` (Boolean) (Net only) If true, the source/destination check is enabled. If false, it is disabled. This value must be false for a NAT VM to perform network address translation (NAT) in a Net.
- `link_public_ip` (Attributes) Information about the public IP association. (see [below for nested schema](#nestedatt--link_public_ip))
- `mac_address` (String) The Media Access Control (MAC) address of the NIC.
- `net_id` (String) The ID of the Net for the NIC.
- `private_dns_name` (String) The name of the private DNS.
- `security_groups` (Attributes List) One or more IDs of security groups for the NIC. (see [below for nested schema](#nestedatt--security_groups))
- `state` (String) The state of the NIC (`available` \| `attaching` \| `in-use` \| `detaching`).
- `subregion_name` (String) The Subregion in which the NIC is located.

<a id="nestedatt--private_ips"></a>
### Nested Schema for `private_ips`

Optional:

- `is_primary` (Boolean) If true, the IP is the primary private IP of the NIC.
- `private_ip` (String) The private IP of the NIC.

Read-Only:

- `link_public_ip` (Attributes) Information about the public IP association. (see [below for nested schema](#nestedatt--private_ips--link_public_ip))
- `private_dns_name` (String) The name of the private DNS.

<a id="nestedatt--private_ips--link_public_ip"></a>
### Nested Schema for `private_ips.link_public_ip`

Read-Only:

- `id` (String) (Required in a Net) The ID representing the association of the public IP with the VM or the NIC.
- `public_dns_name` (String) The name of the public DNS.
- `public_ip` (String) The public IP associated with the NIC.
- `public_ip_account_id` (String) The account ID of the owner of the public IP.
- `public_ip_id` (String) The allocation ID of the public IP.



<a id="nestedatt--link_public_ip"></a>
### Nested Schema for `link_public_ip`

Read-Only:

- `id` (String) (Required in a Net) The ID representing the association of the public IP with the VM or the NIC.
- `public_dns_name` (String) The name of the public DNS.
- `public_ip` (String) The public IP associated with the NIC.
- `public_ip_account_id` (String) The account ID of the owner of the public IP.
- `public_ip_id` (String) The allocation ID of the public IP.


<a id="nestedatt--security_groups"></a>
### Nested Schema for `security_groups`

Read-Only:

- `security_group_id` (String) The ID of the security group.
- `security_group_name` (String) The name of the security group.
