---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_managed_service_bridges Data Source - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_managed_service_bridges (Data Source)



## Example Usage

```terraform
resource "numspot_managed_service_bridge" "managed-service-bridge" {
  source_managed_service_id      = "" // Managed service ID
  destination_managed_service_id = "" // Managed service ID
}

data "numspot_managed_service_bridges" "datasource-managed-service-bridge" {
  depends_on = [numspot_managed_service_bridge.managed-service-bridge.id]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_managed_service_bridges.datasource-managed-service-bridge.items.0.id"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `items` (Attributes List) List of bridges. (see [below for nested schema](#nestedatt--items))

<a id="nestedatt--items"></a>
### Nested Schema for `items`

Read-Only:

- `id` (String)
