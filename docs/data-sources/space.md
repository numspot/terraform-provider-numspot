---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_space Data Source - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_space (Data Source)



## Example Usage

```terraform
resource "numspot_space" "test" {
  organisation_id = "88888888-4444-4444-4444-cccccccccccc"
  name            = "the space"
  description     = "the description"
}

data "numspot_space" "testdata" {
  space_id   = numspot_space.test.id
  depends_on = [numspot_space.test]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_space.testdata.id"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `space_id` (String) Space ID

### Read-Only

- `created_on` (String) Space creation date
- `description` (String) Space description
- `id` (String) Internal ID
- `name` (String) Space name
- `organisation_id` (String) Organisation_id
- `status` (String) status of the space, the space can only be used when the status is ready.
- `updated_on` (String) Space last update