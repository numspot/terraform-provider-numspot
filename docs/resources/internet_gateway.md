---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_internet_gateway Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_internet_gateway (Resource)



## Example Usage

```terraform
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_internet_gateway" "internet_gateway" {
  vpc_id = numspot_vpc.vpc.id
  tags = [
    {
      key   = "name"
      value = "My Internet Gateway"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `tags` (Attributes List) One or more tags associated with the Vpc. (see [below for nested schema](#nestedatt--tags))
- `vpc_id` (String) The ID of the Vpc attached to the Internet gateway.

### Read-Only

- `id` (String) The ID of the Internet gateway.
- `state` (String) The state of the attachment of the Internet gateway to the Vpc (always `available`).

<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Required:

- `key` (String) The key of the tag, with a minimum of 1 character.
- `value` (String) The value of the tag, between 0 and 255 characters.
