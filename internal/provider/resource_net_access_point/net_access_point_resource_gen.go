// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_net_access_point

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func NetAccessPointResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Net access point.",
				MarkdownDescription: "The ID of the Net access point.",
			},
			"net_id": schema.StringAttribute{
				Required:            true,
				Description:         "The ID of the Net.",
				MarkdownDescription: "The ID of the Net.",
			},
			"route_table_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "One or more IDs of route tables to use for the connection.",
				MarkdownDescription: "One or more IDs of route tables to use for the connection.",
			},
			"service_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the service (in the format `com.outscale.region.service`).",
				MarkdownDescription: "The name of the service (in the format `com.outscale.region.service`).",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the Net access point (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
				MarkdownDescription: "The state of the Net access point (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
			},
		},
	}
}

type NetAccessPointModel struct {
	Id            types.String `tfsdk:"id"`
	NetId         types.String `tfsdk:"net_id"`
	RouteTableIds types.List   `tfsdk:"route_table_ids"`
	ServiceName   types.String `tfsdk:"service_name"`
	State         types.String `tfsdk:"state"`
}
