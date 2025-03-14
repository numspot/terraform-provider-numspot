// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package internetgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
)

func InternetGatewayDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Internet gateway.",
							MarkdownDescription: "The ID of the Internet gateway.",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the attachment of the Internet gateway to the Vpc (always `available`).",
							MarkdownDescription: "The state of the attachment of the Internet gateway to the Vpc (always `available`).",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
						"vpc_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Vpc attached to the Internet gateway.",
							MarkdownDescription: "The ID of the Vpc attached to the Internet gateway.",
						},
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more Internet gateways.",
				MarkdownDescription: "Information about one or more Internet gateways.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Internet gateways.",
				MarkdownDescription: "The IDs of the Internet gateways.",
			},
			"link_states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The current states of the attachments between the Internet gateways and the Vpcs (only `available`, if the Internet gateway is attached to a Vpc).",
				MarkdownDescription: "The current states of the attachments between the Internet gateways and the Vpcs (only `available`, if the Internet gateway is attached to a Vpc).",
			},
			"link_vpc_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Vpcs the Internet gateways are attached to.",
				MarkdownDescription: "The IDs of the Vpcs the Internet gateways are attached to.",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the Internet gateways.",
				MarkdownDescription: "The keys of the tags associated with the Internet gateways.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the Internet gateways.",
				MarkdownDescription: "The values of the tags associated with the Internet gateways.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the Internet gateways, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the Internet gateways, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
		},
	}
}

type InternetGatewaysDataSourceModel struct {
	Items      []InternetGatewayModel `tfsdk:"items"`
	IDs        types.List             `tfsdk:"ids"`
	LinkStates types.List             `tfsdk:"link_states"`
	TagKeys    types.List             `tfsdk:"tag_keys"`
	TagValues  types.List             `tfsdk:"tag_values"`
	Tags       types.List             `tfsdk:"tags"`
	LinkVpcIds types.List             `tfsdk:"link_vpc_ids"`
}

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
