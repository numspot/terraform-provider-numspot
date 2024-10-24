---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_virtual_gateway Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_virtual_gateway (Resource)



## Example Usage

```terraform
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "vg" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.vpc.id

  tags = [
    {
      key   = "name"
      value = "My Virtual Gateway"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `connection_type` (String) The type of VPN connection supported by the virtual gateway (only `ipsec.1` is supported).

### Optional

- `tags` (Attributes List) One or more tags associated with the resource. (see [below for nested schema](#nestedatt--tags))
- `vpc_id` (String) The ID of the Vpc to which the virtual gateway is attached.

### Read-Only

- `id` (String) The ID of the virtual gateway.
- `state` (String) The state of the virtual gateway (`pending` \| `available` \| `deleting` \| `deleted`).
- `vpc_to_virtual_gateway_links` (Attributes List) the Vpc to which the virtual gateway is attached. (see [below for nested schema](#nestedatt--vpc_to_virtual_gateway_links))

<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Required:

- `key` (String) The key of the tag, with a minimum of 1 character.
- `value` (String) The value of the tag, between 0 and 255 characters.


<a id="nestedatt--vpc_to_virtual_gateway_links"></a>
### Nested Schema for `vpc_to_virtual_gateway_links`

Read-Only:

- `state` (String) The state of the attachment (`attaching` \| `attached` \| `detaching` \| `detached`).
- `vpc_id` (String) The ID of the Vpc to which the virtual gateway is attached.
