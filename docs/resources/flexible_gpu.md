---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_flexible_gpu Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_flexible_gpu (Resource)



## Example Usage

```terraform
resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_vm" "vm" {
  image_id  = "ami-0b7df82c"
  type      = "ns-cus6-4c8r"
  subnet_id = numspot_subnet.subnet.id
}

resource "numspot_flexible_gpu" "gpu" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "eu-west-2a"
  vm_id                  = numspot_vm.vm.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `availability_zone_name` (String) The Subregion in which you want to create the fGPU.
- `model_name` (String) The model of fGPU you want to allocate.

### Optional

- `generation` (String) The processor generation that the fGPU must be compatible with. If not specified, the oldest possible processor generation is selected (as provided by [ReadFlexibleGpuCatalog](#readflexiblegpucatalog) for the specified model of fGPU).
- `vm_id` (String) The ID of the VM the fGPU is attached to, if any.

### Read-Only

- `delete_on_vm_deletion` (Boolean) If true, the fGPU is deleted when the VM is terminated.
- `id` (String) The ID of the fGPU.
- `state` (String) The state of the fGPU (`allocated` \| `attaching` \| `attached` \| `detaching`).
