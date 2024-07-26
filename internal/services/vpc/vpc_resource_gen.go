// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func VpcResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dhcp_options_set_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Description:         "The ID of the DHCP options set (or `default` if you want to associate the default one).",
				MarkdownDescription: "The ID of the DHCP options set (or `default` if you want to associate the default one).",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Vpc.",
				MarkdownDescription: "The ID of the Vpc.",
			},
			"ip_range": schema.StringAttribute{
				Required:            true,
				Description:         "The IP range for the Vpc, in CIDR notation (for example, `10.0.0.0/16`).",
				MarkdownDescription: "The IP range for the Vpc, in CIDR notation (for example, `10.0.0.0/16`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the Vpc (`pending` \\| `available` \\| `deleting`).",
				MarkdownDescription: "The state of the Vpc (`pending` \\| `available` \\| `deleting`).",
			},
			"tags": tags.TagsSchema(ctx),
			"tenancy": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The tenancy options for the VMs:<br />\n- `default` if a VM created in a Vpc can be launched with any tenancy.<br />\n- `dedicated` if it can be launched with dedicated tenancy VMs running on single-tenant hardware.<br />\n- `dedicated group ID`: if it can be launched in a dedicated group on single-tenant hardware.",
				MarkdownDescription: "The tenancy options for the VMs:<br />\n- `default` if a VM created in a Vpc can be launched with any tenancy.<br />\n- `dedicated` if it can be launched with dedicated tenancy VMs running on single-tenant hardware.<br />\n- `dedicated group ID`: if it can be launched in a dedicated group on single-tenant hardware.",
			},
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated",
	}
}

type VpcModel struct {
	DhcpOptionsSetId types.String `tfsdk:"dhcp_options_set_id"`
	Id               types.String `tfsdk:"id"`
	IpRange          types.String `tfsdk:"ip_range"`
	State            types.String `tfsdk:"state"`
	Tags             types.List   `tfsdk:"tags"`
	Tenancy          types.String `tfsdk:"tenancy"`
}
