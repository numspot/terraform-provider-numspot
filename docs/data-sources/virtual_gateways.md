---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_virtual_gateways Data Source - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_virtual_gateways (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `connection_types` (List of String) The type of VPN connection supported by the virtual gateway (only `ipsec.1` is supported).
- `ids` (List of String) ID for ReadVirtualGateways
- `link_states` (List of String) The state of the attachment (`attaching` \| `attached` \| `detaching` \| `detached`).
- `link_vpc_ids` (List of String) The ID of the Net to which the virtual gateway is attached.
- `states` (List of String) The state of the virtual gateway (`pending` \| `available` \| `deleting` \| `deleted`).
- `tag_keys` (List of String) The keys of the tags associated with the virtual gateways.
- `tag_values` (List of String) The values of the tags associated with the virtual gateways.
- `tags` (List of String) The key/value combination of the tags associated with the virtual gateways, in the following format: "Filters":{"Tags":["TAGKEY=TAGVALUE"]}.

### Read-Only

- `virtual_gateways` (Attributes List) (see [below for nested schema](#nestedatt--virtual_gateways))

<a id="nestedatt--virtual_gateways"></a>
### Nested Schema for `virtual_gateways`

Required:

- `id` (String) ID for ReadVirtualGateways

Optional:

- `tags` (Attributes List) One or more tags associated with the resource. (see [below for nested schema](#nestedatt--virtual_gateways--tags))

Read-Only:

- `connection_type` (String) The type of VPN connection supported by the virtual gateway (only `ipsec.1` is supported).
- `net_to_virtual_gateway_links` (Attributes List) The Net to which the virtual gateway is attached. (see [below for nested schema](#nestedatt--virtual_gateways--net_to_virtual_gateway_links))
- `state` (String) The state of the virtual gateway (`pending` \| `available` \| `deleting` \| `deleted`).

<a id="nestedatt--virtual_gateways--tags"></a>
### Nested Schema for `virtual_gateways.tags`

Required:

- `key` (String) The key of the tag, with a minimum of 1 character.
- `value` (String) The value of the tag, between 0 and 255 characters.


<a id="nestedatt--virtual_gateways--net_to_virtual_gateway_links"></a>
### Nested Schema for `virtual_gateways.net_to_virtual_gateway_links`

Read-Only:

- `state` (String) The state of the attachment (`attaching` \| `attached` \| `detaching` \| `detached`).
- `vpc_id` (String) The ID of the Net to which the virtual gateway is attached.