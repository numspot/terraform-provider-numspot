package space

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SpaceResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the space.",
				MarkdownDescription: "The ID of the space.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Space name",
				MarkdownDescription: "Space name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:            true,
				Description:         "Space description",
				MarkdownDescription: "Space description",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organisation_id": schema.StringAttribute{
				Required:            true,
				Description:         "Organisation ID",
				MarkdownDescription: "Organisation ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				Computed:            true,
				Description:         "Space ID",
				MarkdownDescription: "Space ID",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "status of the space, the space can only be used when the status is ready. Enum: \"QUEUED\" \"RUNNING\" \"READY\" \"FAILED\"",
				MarkdownDescription: "status of the space, the space can only be used when the status is ready. Enum: \"QUEUED\" \"RUNNING\" \"READY\" \"FAILED\"",
			},
			"created_on": schema.StringAttribute{
				Computed:            true,
				Description:         "Space creation date.",
				MarkdownDescription: "Space creation date.",
			},
			"updated_on": schema.StringAttribute{
				Computed:            true,
				Description:         "Space last update.",
				MarkdownDescription: "Space last update.",
			},
		},
	}
}

type SpaceModel struct {
	CreatedOn      types.String `tfsdk:"created_on"`
	Description    types.String `tfsdk:"description"`
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganisationId types.String `tfsdk:"organisation_id"`
	SpaceId        types.String `tfsdk:"space_id"`
	Status         types.String `tfsdk:"status"`
	UpdatedOn      types.String `tfsdk:"updated_on"`
}
