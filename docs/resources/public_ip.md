---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_public_ip Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_public_ip (Resource)



## Example Usage

```terraform
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id                 = numspot_vpc.vpc.id
  ip_range               = "10.101.1.0/24"
  availability_zone_name = "cloudgouv-eu-west-1a"
}

resource "numspot_vm" "vm" {
  image_id  = "ami-12345678"
  type      = "ns-cus6-2c4r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_internet_gateway" "ig" {
  vpc_id = numspot_vpc.vpc.id
}

resource "numspot_public_ip" "public_ip" {
  vm_id = numspot_vm.vm.id
  tags = [
    {
      key   = "name"
      value = "Terraform-Test-PublicIp"
    }
  ]
  depends_on = [numspot_internet_gateway.ig]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `nic_id` (String) The ID of the NIC the public IP is associated with (if any).
- `tags` (Attributes List) One or more tags associated with the resource. (see [below for nested schema](#nestedatt--tags))
- `vm_id` (String) The ID of the VM the public IP is associated with (if any).

### Read-Only

- `id` (String) The allocation ID of the public IP.
- `link_public_ip` (String) The ID of the association between the public IP and VM/NIC (if any).
- `link_public_ip_id` (String) (Required in a Vpc) The ID representing the association of the public IP with the VM or the NIC.
- `private_ip` (String) The private IP associated with the public IP.
- `public_ip` (String) The public IP.

<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Required:

- `key` (String) The key of the tag, with a minimum of 1 character.
- `value` (String) The value of the tag, between 0 and 255 characters.