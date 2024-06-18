// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_flexible_gpu

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func FlexibleGpuResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone_name": schema.StringAttribute{
				Required:            true,
				Description:         "The Subregion in which you want to create the fGPU.",
				MarkdownDescription: "The Subregion in which you want to create the fGPU.",
			},
			"delete_on_vm_deletion": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "If true, the fGPU is deleted when the VM is terminated.",
				MarkdownDescription: "If true, the fGPU is deleted when the VM is terminated.",
				Default:             booldefault.StaticBool(false),
			},
			"generation": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The processor generation that the fGPU must be compatible with. If not specified, the oldest possible processor generation is selected (as provided by [ReadFlexibleGpuCatalog](#readflexiblegpucatalog) for the specified model of fGPU).",
				MarkdownDescription: "The processor generation that the fGPU must be compatible with. If not specified, the oldest possible processor generation is selected (as provided by [ReadFlexibleGpuCatalog](#readflexiblegpucatalog) for the specified model of fGPU).",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the fGPU.",
				MarkdownDescription: "The ID of the fGPU.",
			},
			"model_name": schema.StringAttribute{
				Required:            true,
				Description:         "The model of fGPU you want to allocate. For more information, see [About Flexible GPUs](https://docs.outscale.com/en/userguide/About-Flexible-GPUs.html).",
				MarkdownDescription: "The model of fGPU you want to allocate. For more information, see [About Flexible GPUs](https://docs.outscale.com/en/userguide/About-Flexible-GPUs.html).",
			},
			"space_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "space identifier",
				MarkdownDescription: "space identifier",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the fGPU (`allocated` \\| `attaching` \\| `attached` \\| `detaching`).",
				MarkdownDescription: "The state of the fGPU (`allocated` \\| `attaching` \\| `attached` \\| `detaching`).",
			},
			"vm_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Description:         "The ID of the VM the fGPU is attached to, if any.",
				MarkdownDescription: "The ID of the VM the fGPU is attached to, if any.",
			},
		},
	}
}

type FlexibleGpuModel struct {
	AvailabilityZoneName types.String `tfsdk:"availability_zone_name"`
	DeleteOnVmDeletion   types.Bool   `tfsdk:"delete_on_vm_deletion"`
	Generation           types.String `tfsdk:"generation"`
	Id                   types.String `tfsdk:"id"`
	ModelName            types.String `tfsdk:"model_name"`
	SpaceId              types.String `tfsdk:"space_id"`
	State                types.String `tfsdk:"state"`
	VmId                 types.String `tfsdk:"vm_id"`
}
