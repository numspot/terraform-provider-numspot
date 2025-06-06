---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_snapshot Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_snapshot (Resource)



## Example Usage

```terraform
resource "numspot_volume" "volume" {
  type                   = "standard"
  size                   = 11
  availability_zone_name = "eu-west-2a"
}

resource "numspot_snapshot" "snapshot" {
  volume_id   = numspot_volume.volume.id
  description = "A beautiful snapshot"
  tags = [
    {
      key   = "name"
      value = "My Snapshot"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `description` (String) A description for the snapshot.
- `source_region_name` (String) **(when copying a snapshot)** The name of the source Region, which must be the same as the Region of your account.
- `source_snapshot_id` (String) **(when copying a snapshot)** The ID of the snapshot you want to copy.
- `tags` (Attributes List) One or more tags associated with the snapshot. (see [below for nested schema](#nestedatt--tags))
- `volume_id` (String) **(when creating from a volume)** The ID of the volume you want to create a snapshot of.

### Read-Only

- `access` (Attributes) Permissions for the resource. (see [below for nested schema](#nestedatt--access))
- `creation_date` (String) The date and time of creation of the snapshot.
- `id` (String) The ID of the snapshot.
- `progress` (Number) The progress of the snapshot, as a percentage.
- `state` (String) The state of the snapshot (`in-queue` \| `completed` \| `error`).
- `volume_size` (Number) The size of the volume used to create the snapshot, in gibibytes (GiB).

<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Required:

- `key` (String) The key of the tag, with a minimum of 1 character.
- `value` (String) The value of the tag, between 0 and 255 characters.


<a id="nestedatt--access"></a>
### Nested Schema for `access`

Read-Only:

- `is_public` (Boolean) A global permission for all accounts.<br />
(Request) Set this parameter to true to make the resource public (if the parent parameter is `Additions`) or to make the resource private (if the parent parameter is `Removals`).<br />
(Response) If true, the resource is public. If false, the resource is private.
