---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_nat_gateway Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_nat_gateway (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `public_ip_id` (String) The allocation ID of the public IP to associate with the NAT service.<br />
If the public IP is already associated with another resource, you must first disassociate it.
- `subnet_id` (String) The ID of the Subnet in which you want to create the NAT service.

### Read-Only

- `id` (String) The ID of the NAT service.
- `public_ips` (Attributes List) Information about the public IP or IPs associated with the NAT service. (see [below for nested schema](#nestedatt--public_ips))
- `state` (String) The state of the NAT service (`pending` \| `available` \| `deleting` \| `deleted`).
- `vpc_id` (String) The ID of the Net in which the NAT service is.

<a id="nestedatt--public_ips"></a>
### Nested Schema for `public_ips`

Read-Only:

- `public_ip` (String) The public IP associated with the NAT service.
- `public_ip_id` (String) The allocation ID of the public IP associated with the NAT service.