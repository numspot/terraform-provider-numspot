// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package snapshot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
)

func SnapshotDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"access": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"is_public": schema.BoolAttribute{
									Computed:            true,
									Description:         "A global permission for all accounts.<br />\n(Request) Set this parameter to true to make the resource public (if the parent parameter is `Additions`) or to make the resource private (if the parent parameter is `Removals`).<br />\n(Response) If true, the resource is public. If false, the resource is private.",
									MarkdownDescription: "A global permission for all accounts.<br />\n(Request) Set this parameter to true to make the resource public (if the parent parameter is `Additions`) or to make the resource private (if the parent parameter is `Removals`).<br />\n(Response) If true, the resource is public. If false, the resource is private.",
								},
							},
							CustomType: AccessType{
								ObjectType: types.ObjectType{
									AttrTypes: AccessValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Permissions for the resource.",
							MarkdownDescription: "Permissions for the resource.",
						},
						"creation_date": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time of creation of the snapshot.",
							MarkdownDescription: "The date and time of creation of the snapshot.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							Description:         "The description of the snapshot.",
							MarkdownDescription: "The description of the snapshot.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the snapshot.",
							MarkdownDescription: "The ID of the snapshot.",
						},
						"progress": schema.Int64Attribute{
							Computed:            true,
							Description:         "The progress of the snapshot, as a percentage.",
							MarkdownDescription: "The progress of the snapshot, as a percentage.",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the snapshot (`in-queue` \\| `completed` \\| `error`).",
							MarkdownDescription: "The state of the snapshot (`in-queue` \\| `completed` \\| `error`).",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
						"volume_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the volume used to create the snapshot.",
							MarkdownDescription: "The ID of the volume used to create the snapshot.",
						},
						"volume_size": schema.Int64Attribute{
							Computed:            true,
							Description:         "The size of the volume used to create the snapshot, in gibibytes (GiB).",
							MarkdownDescription: "The size of the volume used to create the snapshot, in gibibytes (GiB).",
						},
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more snapshots and their permissions.",
				MarkdownDescription: "Information about one or more snapshots and their permissions.",
			},
			"descriptions": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The descriptions of the snapshots.",
				MarkdownDescription: "The descriptions of the snapshots.",
			},
			"from_creation_date": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The beginning of the time period, in ISO 8601 date-time format (for example, `2020-06-14T00:00:00.000Z`).",
				MarkdownDescription: "The beginning of the time period, in ISO 8601 date-time format (for example, `2020-06-14T00:00:00.000Z`).",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the snapshots.",
				MarkdownDescription: "The IDs of the snapshots.",
			},
			"is_public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "If true, lists all public volumes. If false, lists all private volumes.",
				MarkdownDescription: "If true, lists all public volumes. If false, lists all private volumes.",
			},
			"progresses": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The progresses of the snapshots, as a percentage.",
				MarkdownDescription: "The progresses of the snapshots, as a percentage.",
			},
			"states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the snapshots (`in-queue` \\| `completed` \\| `error`).",
				MarkdownDescription: "The states of the snapshots (`in-queue` \\| `completed` \\| `error`).",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the snapshots.",
				MarkdownDescription: "The keys of the tags associated with the snapshots.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the snapshots.",
				MarkdownDescription: "The values of the tags associated with the snapshots.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the snapshots, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the snapshots, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
			"to_creation_date": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The end of the time period, in ISO 8601 date-time format (for example, `2020-06-30T00:00:00.000Z`).",
				MarkdownDescription: "The end of the time period, in ISO 8601 date-time format (for example, `2020-06-30T00:00:00.000Z`).",
			},
			"volume_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the volumes used to create the snapshots.",
				MarkdownDescription: "The IDs of the volumes used to create the snapshots.",
			},
			"volume_sizes": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The sizes of the volumes used to create the snapshots, in gibibytes (GiB).",
				MarkdownDescription: "The sizes of the volumes used to create the snapshots, in gibibytes (GiB).",
			},
			// MANUALLY EDITED : SpaceId removed
		},
	}
}

type SnapshotModelDatasource struct { // MANUALLY EDITED : Create Model from ItemsValue struct
	Access       AccessValue  `tfsdk:"access"`
	CreationDate types.String `tfsdk:"creation_date"`
	Description  types.String `tfsdk:"description"`
	Id           types.String `tfsdk:"id"`
	Progress     types.Int64  `tfsdk:"progress"`
	State        types.String `tfsdk:"state"`
	Tags         types.List   `tfsdk:"tags"`
	VolumeId     types.String `tfsdk:"volume_id"`
	VolumeSize   types.Int64  `tfsdk:"volume_size"`
}

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed