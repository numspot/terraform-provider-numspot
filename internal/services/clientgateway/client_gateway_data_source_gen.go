// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package clientgateway

import (
	"context"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ClientGatewayDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"bgp_asn": schema.Int64Attribute{
							Computed:            true,
							Description:         "The Autonomous System Number (ASN) used by the Border Gateway Protocol (BGP) to find the path to your client gateway through the Internet.",
							MarkdownDescription: "The Autonomous System Number (ASN) used by the Border Gateway Protocol (BGP) to find the path to your client gateway through the Internet.",
						},
						"connection_type": schema.StringAttribute{
							Computed:            true,
							Description:         "The type of communication tunnel used by the client gateway (only `ipsec.1` is supported).",
							MarkdownDescription: "The type of communication tunnel used by the client gateway (only `ipsec.1` is supported).",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the client gateway.",
							MarkdownDescription: "The ID of the client gateway.",
						},
						"public_ip": schema.StringAttribute{
							Computed:            true,
							Description:         "The public IPv4 address of the client gateway (must be a fixed address into a NATed network).",
							MarkdownDescription: "The public IPv4 address of the client gateway (must be a fixed address into a NATed network).",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the client gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
							MarkdownDescription: "The state of the client gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more client gateways.",
				MarkdownDescription: "Information about one or more client gateways.",
			},
			"bgp_asns": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The Border Gateway Protocol (BGP) Autonomous System Numbers (ASNs) of the connections.",
				MarkdownDescription: "The Border Gateway Protocol (BGP) Autonomous System Numbers (ASNs) of the connections.",
			},
			"connection_types": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The types of communication tunnels used by the client gateways (only `ipsec.1` is supported).",
				MarkdownDescription: "The types of communication tunnels used by the client gateways (only `ipsec.1` is supported).",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the client gateways.",
				MarkdownDescription: "The IDs of the client gateways.",
			}, "public_ips": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The public IPv4 addresses of the client gateways.",
				MarkdownDescription: "The public IPv4 addresses of the client gateways.",
			},
			"states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the client gateways (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
				MarkdownDescription: "The states of the client gateways (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the client gateways.",
				MarkdownDescription: "The keys of the tags associated with the client gateways.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the client gateways.",
				MarkdownDescription: "The values of the tags associated with the client gateways.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the client gateways, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the client gateways, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
			// MANUALLY EDITED : SpaceId Removed
		},
	}
}

// MANUALLY EDITED : Model declaration removed

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
