---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_net Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_net (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `ip_range` (String) The IP range for the Net, in CIDR notation (for example, `10.0.0.0/16`).

### Optional

- `tenancy` (String) The tenancy options for the VMs (`default` if a VM created in a Net can be launched with any tenancy, `dedicated` if it can be launched with dedicated tenancy VMs running on single-tenant hardware).

### Read-Only

- `dhcp_options_set_id` (String) The ID of the DHCP options set (or `default` if you want to associate the default one).
- `id` (String) The ID of the Net.
- `state` (String) The state of the Net (`pending` \| `available` \| `deleted`).