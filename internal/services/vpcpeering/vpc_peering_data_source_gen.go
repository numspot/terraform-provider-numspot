// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package vpcpeering

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func VpcPeeringDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"accepter_vpc_ip_ranges": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IP ranges of the peer Vpcs, in CIDR notation (for example, `10.0.0.0/24`).",
				MarkdownDescription: "The IP ranges of the peer Vpcs, in CIDR notation (for example, `10.0.0.0/24`).",
			},
			"accepter_vpc_vpc_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the peer Vpcs",
				MarkdownDescription: "The IDs of the peer Vpcs",
			},
			"expiration_dates": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The dates and times at which the Vpc peerings expire, in ISO 8601 date-time format (for example, `2020-06-14T00:00:00.000Z`).",
				MarkdownDescription: "The dates and times at which the Vpc peerings expire, in ISO 8601 date-time format (for example, `2020-06-14T00:00:00.000Z`).",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Vpc peerings.",
				MarkdownDescription: "The IDs of the Vpc peerings.",
			},
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"accepter_vpc": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"ip_range": schema.StringAttribute{
									Computed:            true,
									Description:         "The IP range for the accepter Net, in CIDR notation (for example, `10.0.0.0/16`).",
									MarkdownDescription: "The IP range for the accepter Net, in CIDR notation (for example, `10.0.0.0/16`).",
								},
								"vpc_id": schema.StringAttribute{
									Computed:            true,
									Description:         "The ID of the accepter Vpc.",
									MarkdownDescription: "The ID of the accepter Vpc.",
								},
							},
							CustomType: AccepterVpcType{
								ObjectType: types.ObjectType{
									AttrTypes: AccepterVpcValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the accepter Vpc.",
							MarkdownDescription: "Information about the accepter Vpc.",
						},
						"expiration_date": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time at which the Vpc peerings expire.",
							MarkdownDescription: "The date and time at which the Vpc peerings expire.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Vpc peering.",
							MarkdownDescription: "The ID of the Vpc peering.",
						},
						"source_vpc": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"ip_range": schema.StringAttribute{
									Computed:            true,
									Description:         "The IP range for the source Net, in CIDR notation (for example, `10.0.0.0/16`).",
									MarkdownDescription: "The IP range for the source Net, in CIDR notation (for example, `10.0.0.0/16`).",
								},
								"vpc_id": schema.StringAttribute{
									Computed:            true,
									Description:         "The ID of the source Vpc.",
									MarkdownDescription: "The ID of the source Vpc.",
								},
							},
							CustomType: SourceVpcType{
								ObjectType: types.ObjectType{
									AttrTypes: SourceVpcValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the source Vpc.",
							MarkdownDescription: "Information about the source Vpc.",
						},
						"state": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"message": schema.StringAttribute{
									Computed:            true,
									Description:         "Additional information about the state of the Vpc peering.",
									MarkdownDescription: "Additional information about the state of the Vpc peering.",
								},
								"name": schema.StringAttribute{
									Computed:            true,
									Description:         "The state of the Vpc peering (`pending-acceptance` \\| `active` \\| `rejected` \\| `failed` \\| `expired` \\| `deleted`).",
									MarkdownDescription: "The state of the Vpc peering (`pending-acceptance` \\| `active` \\| `rejected` \\| `failed` \\| `expired` \\| `deleted`).",
								},
							},
							CustomType: StateType{
								ObjectType: types.ObjectType{
									AttrTypes: StateValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the state of the Vpc peering.",
							MarkdownDescription: "Information about the state of the Vpc peering.",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more Vpc peerings.",
				MarkdownDescription: "Information about one or more Vpc peerings.",
			},
			"source_vpc_ip_ranges": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IP ranges of the peer Vpcs",
				MarkdownDescription: "The IP ranges of the peer Vpcs",
			},
			"source_vpc_vpc_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the peer Vpcs",
				MarkdownDescription: "The IDs of the peer Vpcs",
			},
			"state_messages": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Additional information about the states of the Vpc peerings.",
				MarkdownDescription: "Additional information about the states of the Vpc peerings.",
			},
			"state_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the Vpc peerings (`pending-acceptance` \\| `active` \\| `rejected` \\| `failed` \\| `expired` \\| `deleted`).",
				MarkdownDescription: "The states of the Vpc peerings (`pending-acceptance` \\| `active` \\| `rejected` \\| `failed` \\| `expired` \\| `deleted`).",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the Vpc peerings.",
				MarkdownDescription: "The keys of the tags associated with the Vpc peerings.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the Vpc peerings.",
				MarkdownDescription: "The values of the tags associated with the Vpc peerings.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the Vpc peerings, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the Vpc peerings, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
			// MANUALLY EDITED : spaceId removed
		},
	}
}

type VpcPeeringDatasourceItemModel struct { // MANUALLY EDITED : Create Model from ItemsValue struct
	AccepterVpc    AccepterVpcValue `tfsdk:"accepter_vpc"`
	ExpirationDate types.String     `tfsdk:"expiration_date"`
	Id             types.String     `tfsdk:"id"`
	SourceVpc      SourceVpcValue   `tfsdk:"source_vpc"`
	State          StateValue       `tfsdk:"state"`
	Tags           types.List       `tfsdk:"tags"`
}

// MANUALLY EDITED : Model declaration removed

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
