---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_direct_link Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_direct_link (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `bandwidth` (String) The bandwidth of the DirectLink (`1Gbps` \| `10Gbps`).
- `direct_link_name` (String) The name of the DirectLink.
- `location` (String) The code of the requested location for the DirectLink, returned by the [ReadLocations](#readlocations) method.

### Read-Only

- `id` (String) The ID of the DirectLink (for example, `dxcon-xxxxxxxx`).
- `name` (String) The name of the DirectLink.
- `region_name` (String) The Region in which the DirectLink has been created.
- `state` (String) The state of the DirectLink.<br />
* `requested`: The DirectLink is requested but the request has not been validated yet.<br />
* `pending`: The DirectLink request has been validated. It remains in the `pending` state until you establish the physical link.<br />
* `available`: The physical link is established and the connection is ready to use.<br />
 * `deleting`: The deletion process is in progress.<br />
* `deleted`: The DirectLink is deleted.