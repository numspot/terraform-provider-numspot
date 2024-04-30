// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_client_gateway

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ClientGatewayResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bgp_asn": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description:         "The Autonomous System Number (ASN) used by the Border Gateway Protocol (BGP) to find the path to your client gateway through the Internet. This number must be between `1` and `4294967295`.",
				MarkdownDescription: "The Autonomous System Number (ASN) used by the Border Gateway Protocol (BGP) to find the path to your client gateway through the Internet. This number must be between `1` and `4294967295`.",
			},
			"connection_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description:         "The communication protocol used to establish tunnel with your client gateway (only `ipsec.1` is supported).",
				MarkdownDescription: "The communication protocol used to establish tunnel with your client gateway (only `ipsec.1` is supported).",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the client gateway.",
				MarkdownDescription: "The ID of the client gateway.",
			},
			"public_ip": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description:         "The public fixed IPv4 address of your client gateway.",
				MarkdownDescription: "The public fixed IPv4 address of your client gateway.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the client gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
				MarkdownDescription: "The state of the client gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
			},
			"tags": tags.TagsSchema(ctx),
		},
	}
}

type ClientGatewayModel struct {
	BgpAsn         types.Int64  `tfsdk:"bgp_asn"`
	ConnectionType types.String `tfsdk:"connection_type"`
	Id             types.String `tfsdk:"id"`
	PublicIp       types.String `tfsdk:"public_ip"`
	State          types.String `tfsdk:"state"`
	Tags		   types.List	`tfsdk:"tags"`
}
