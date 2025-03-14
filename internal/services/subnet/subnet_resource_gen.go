// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package subnet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func SubnetResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The name of the Subregion in which you want to create the Subnet.",
				MarkdownDescription: "The name of the Subregion in which you want to create the Subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"available_ips_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of available IPs in the Subnets.",
				MarkdownDescription: "The number of available IPs in the Subnets.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Subnet.",
				MarkdownDescription: "The ID of the Subnet.",
			},
			"ip_range": schema.StringAttribute{
				Required:            true,
				Description:         "The IP range in the Subnet, in CIDR notation (for example, `10.0.0.0/16`).<br />\nThe IP range of the Subnet can be either the same as the Vpc one if you create only a single Subnet in this Net, or a subset of the Vpc one. In case of several Subnets in a Vpc, their IP ranges must not overlap. The smallest Subnet you can create uses a /29 netmask (eight IPs).",
				MarkdownDescription: "The IP range in the Subnet, in CIDR notation (for example, `10.0.0.0/16`).<br />\nThe IP range of the Subnet can be either the same as the Vpc one if you create only a single Subnet in this Net, or a subset of the Vpc one. In case of several Subnets in a Vpc, their IP ranges must not overlap. The smallest Subnet you can create uses a /29 netmask (eight IPs).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"map_public_ip_on_launch": schema.BoolAttribute{
				Computed:            true,
				Optional:            true, // MANUALLY EDITED : Add Optional attribute
				Description:         "If true, a public IP is assigned to the network interface cards (NICs) created in the specified Subnet.",
				MarkdownDescription: "If true, a public IP is assigned to the network interface cards (NICs) created in the specified Subnet.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the Subnet (`pending` \\| `available` \\| `deleted`).",
				MarkdownDescription: "The state of the Subnet (`pending` \\| `available` \\| `deleted`).",
			},
			"vpc_id": schema.StringAttribute{
				Required:            true,
				Description:         "The ID of the Vpc for which you want to create a Subnet.",
				MarkdownDescription: "The ID of the Vpc for which you want to create a Subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
			// MANUALLY EDITED : SpaceId removed
		},
	}
}

type SubnetModel struct {
	AvailabilityZoneName types.String `tfsdk:"availability_zone_name"`
	AvailableIpsCount    types.Int64  `tfsdk:"available_ips_count"`
	Id                   types.String `tfsdk:"id"`
	IpRange              types.String `tfsdk:"ip_range"`
	MapPublicIpOnLaunch  types.Bool   `tfsdk:"map_public_ip_on_launch"`
	State                types.String `tfsdk:"state"`
	Tags                 types.List   `tfsdk:"tags"`
	VpcId                types.String `tfsdk:"vpc_id"`
	// MANUALLY EDITED : SpaceId removed
}

// MANUALLY EDITED : removed functions related to Tags
