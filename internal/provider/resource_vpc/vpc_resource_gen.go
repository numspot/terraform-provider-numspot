// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_vpc

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"

	//"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func VpcResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dhcp_options_set_id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the DHCP options set (or `default` if you want to associate the default one).",
				MarkdownDescription: "The ID of the DHCP options set (or `default` if you want to associate the default one).",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Net.",
				MarkdownDescription: "The ID of the Net.",
			},
			"ip_range": schema.StringAttribute{
				Required:            true,
				Description:         "The IP range for the Net, in CIDR notation (for example, `10.0.0.0/16`).",
				MarkdownDescription: "The IP range for the Net, in CIDR notation (for example, `10.0.0.0/16`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the Net (`pending` \\| `available` \\| `deleting`).",
				MarkdownDescription: "The state of the Net (`pending` \\| `available` \\| `deleting`).",
			},
			"tenancy": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The tenancy options for the VMs:<br />\n- `default` if a VM created in a Net can be launched with any tenancy.<br />\n- `dedicated` if it can be launched with dedicated tenancy VMs running on single-tenant hardware.<br />\n- `dedicated group ID`: if it can be launched in a dedicated group on single-tenant hardware.",
				MarkdownDescription: "The tenancy options for the VMs:<br />\n- `default` if a VM created in a Net can be launched with any tenancy.<br />\n- `dedicated` if it can be launched with dedicated tenancy VMs running on single-tenant hardware.<br />\n- `dedicated group ID`: if it can be launched in a dedicated group on single-tenant hardware.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": tags.TagsSchema(ctx),
		},
	}
}

type VpcModel struct {
	DhcpOptionsSetId types.String `tfsdk:"dhcp_options_set_id"`
	Id               types.String `tfsdk:"id"`
	IpRange          types.String `tfsdk:"ip_range"`
	State            types.String `tfsdk:"state"`
	Tenancy          types.String `tfsdk:"tenancy"`
	Tags             types.List   `tfsdk:"tags"`
}
