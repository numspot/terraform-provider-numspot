// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package clientgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"

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
				Description:         "The Autonomous System Number (ASN) used by the Border Gateway Protocol (BGP) to find the path to your client gateway through the Internet. <br/>\nThis number must be between `1` and `4294967295`. If you do not have an ASN, you can choose one between 64512 and 65534, or between 4200000000 and 4294967294.",
				MarkdownDescription: "The Autonomous System Number (ASN) used by the Border Gateway Protocol (BGP) to find the path to your client gateway through the Internet. <br/>\nThis number must be between `1` and `4294967295`. If you do not have an ASN, you can choose one between 64512 and 65534, or between 4200000000 and 4294967294.",
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
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated",
	}
}

type ClientGatewayModel struct {
	BgpAsn         types.Int64  `tfsdk:"bgp_asn"`
	ConnectionType types.String `tfsdk:"connection_type"`
	Id             types.String `tfsdk:"id"`
	PublicIp       types.String `tfsdk:"public_ip"`
	State          types.String `tfsdk:"state"`
	Tags           types.List   `tfsdk:"tags"`
}
