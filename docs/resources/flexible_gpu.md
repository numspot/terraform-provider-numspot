---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_flexible_gpu Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_flexible_gpu (Resource)



## Example Usage

```terraform
resource "numspot_vm" "vm" {
  image_id = "ami-026ce760"
  type     = "ns-mus6-2c16r"
}

resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "cloudgouv-eu-west-1a"
  delete_on_vm_deletion  = true
  vm_id                  = numspot_vm.vm.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `availability_zone_name` (String) The Subregion in which you want to create the fGPU.
- `model_name` (String) The model of fGPU you want to allocate. For more information, see [About Flexible GPUs](https://docs.outscale.com/en/userguide/About-Flexible-GPUs.html).

### Optional

- `delete_on_vm_deletion` (Boolean) If true, the fGPU is deleted when the VM is terminated.
- `generation` (String) The processor generation that the fGPU must be compatible with. If not specified, the oldest possible processor generation is selected (as provided by [ReadFlexibleGpuCatalog](#readflexiblegpucatalog) for the specified model of fGPU).
- `space_id` (String) space identifier
- `vm_id` (String) The ID of the VM the fGPU is attached to, if any.

### Read-Only

- `id` (String) The ID of the fGPU.
- `state` (String) The state of the fGPU (`allocated` \| `attaching` \| `attached` \| `detaching`).
