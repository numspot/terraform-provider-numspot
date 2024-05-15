package resource_space

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
				Description:         "Space description",
				MarkdownDescription: "Space description",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	OrganisationId types.String `tfsdk:"organisation_id"`
	Status         types.String `tfsdk:"status"`
	CreatedOn      types.String `tfsdk:"created_on"`
	UpdatedOn      types.String `tfsdk:"updated_on"`
}
