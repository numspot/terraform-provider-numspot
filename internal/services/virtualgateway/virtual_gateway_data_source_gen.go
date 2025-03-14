// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package virtualgateway

import (
	"context"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func VirtualGatewayDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"connection_type": schema.StringAttribute{
							Computed:            true,
							Description:         "The type of VPN connection supported by the virtual gateway (only `ipsec.1` is supported).",
							MarkdownDescription: "The type of VPN connection supported by the virtual gateway (only `ipsec.1` is supported).",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the virtual gateway.",
							MarkdownDescription: "The ID of the virtual gateway.",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the virtual gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
							MarkdownDescription: "The state of the virtual gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
						"vpc_to_virtual_gateway_links": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"state": schema.StringAttribute{
										Computed:            true,
										Description:         "The state of the attachment (`attaching` \\| `attached` \\| `detaching` \\| `detached`).",
										MarkdownDescription: "The state of the attachment (`attaching` \\| `attached` \\| `detaching` \\| `detached`).",
									},
									"vpc_id": schema.StringAttribute{
										Computed:            true,
										Description:         "The ID of the Vpc to which the virtual gateway is attached.",
										MarkdownDescription: "The ID of the Vpc to which the virtual gateway is attached.",
									},
								},
								CustomType: VpcToVirtualGatewayLinksType{
									ObjectType: types.ObjectType{
										AttrTypes: VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "the Vpc to which the virtual gateway is attached.",
							MarkdownDescription: "the Vpc to which the virtual gateway is attached.",
						},
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more virtual gateways.",
				MarkdownDescription: "Information about one or more virtual gateways.",
			},
			"connection_types": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The types of the virtual gateways (only `ipsec.1` is supported).",
				MarkdownDescription: "The types of the virtual gateways (only `ipsec.1` is supported).",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the virtual gateways.",
				MarkdownDescription: "The IDs of the virtual gateways.",
			},
			"link_states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The current states of the attachments between the virtual gateways and the Vpcs (`attaching` \\| `attached` \\| `detaching` \\| `detached`).",
				MarkdownDescription: "The current states of the attachments between the virtual gateways and the Vpcs (`attaching` \\| `attached` \\| `detaching` \\| `detached`).",
			},
			"link_vpc_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Vpcs the virtual gateways are attached to.",
				MarkdownDescription: "The IDs of the Vpcs the virtual gateways are attached to.",
			},
			"states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the virtual gateways (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
				MarkdownDescription: "The states of the virtual gateways (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the virtual gateways.",
				MarkdownDescription: "The keys of the tags associated with the virtual gateways.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the virtual gateways.",
				MarkdownDescription: "The values of the tags associated with the virtual gateways.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the virtual gateways, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the virtual gateways, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
			// MANUALLY EDITED : remove spaceId
		},
	}
}

type VirtualGatewayModelItemDataSource struct {
	ConnectionType           types.String `tfsdk:"connection_type"`
	Id                       types.String `tfsdk:"id"`
	State                    types.String `tfsdk:"state"`
	Tags                     types.List   `tfsdk:"tags"`
	VpcToVirtualGatewayLinks types.List   `tfsdk:"vpc_to_virtual_gateway_links"`
	// MANUALLY EDITED : SpaceId Removed
}

// MANUALLY EDITED : Model declaration removed

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
