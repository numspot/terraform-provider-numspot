// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package directlink

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DirectLinkResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bandwidth": schema.StringAttribute{
				Required:            true,
				Description:         "The bandwidth of the DirectLink (`1Gbps` \\| `10Gbps`).",
				MarkdownDescription: "The bandwidth of the DirectLink (`1Gbps` \\| `10Gbps`).",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the DirectLink (for example, `dxcon-xxxxxxxx`).",
				MarkdownDescription: "The ID of the DirectLink (for example, `dxcon-xxxxxxxx`).",
			},
			"location": schema.StringAttribute{
				Required:            true,
				Description:         "The code of the requested location for the DirectLink, returned by the [ReadLocations](#readlocations) method.",
				MarkdownDescription: "The code of the requested location for the DirectLink, returned by the [ReadLocations](#readlocations) method.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the DirectLink.",
				MarkdownDescription: "The name of the DirectLink.",
			},
			"region_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The Region in which the DirectLink has been created.",
				MarkdownDescription: "The Region in which the DirectLink has been created.",
			},
			"space_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "space identifier",
				MarkdownDescription: "space identifier",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the DirectLink.<br />\n* `requested`: The DirectLink is requested but the request has not been validated yet.<br />\n* `pending`: The DirectLink request has been validated. It remains in the `pending` state until you establish the physical link.<br />\n* `available`: The physical link is established and the connection is ready to use.<br />\n * `deleting`: The deletion process is in progress.<br />\n* `deleted`: The DirectLink is deleted.",
				MarkdownDescription: "The state of the DirectLink.<br />\n* `requested`: The DirectLink is requested but the request has not been validated yet.<br />\n* `pending`: The DirectLink request has been validated. It remains in the `pending` state until you establish the physical link.<br />\n* `available`: The physical link is established and the connection is ready to use.<br />\n * `deleting`: The deletion process is in progress.<br />\n* `deleted`: The DirectLink is deleted.",
			},
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated",
	}
}

type DirectLinkModel struct {
	Bandwidth  types.String `tfsdk:"bandwidth"`
	Id         types.String `tfsdk:"id"`
	Location   types.String `tfsdk:"location"`
	Name       types.String `tfsdk:"name"`
	RegionName types.String `tfsdk:"region_name"`
	SpaceId    types.String `tfsdk:"space_id"`
	State      types.String `tfsdk:"state"`
}
