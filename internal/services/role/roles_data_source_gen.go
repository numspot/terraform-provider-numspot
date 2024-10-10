// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package role

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func RolesDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "Internal ID",
							MarkdownDescription: "Internal ID",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "Human-readable name",
							MarkdownDescription: "Human-readable name",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							Description:         "Human-readable description",
							MarkdownDescription: "Human-readable description",
						},
						"created_on": schema.StringAttribute{
							Computed:            true,
							Description:         "creation date",
							MarkdownDescription: "creation date",
						},
						"updated_on": schema.StringAttribute{
							Computed:            true,
							Description:         "last update",
							MarkdownDescription: "last update",
						},
					},
				}, // MANUALLY EDITED : Removed CustomType block
			},
			"space_id": schema.StringAttribute{
				Required:            true,
				Description:         "Space ID",
				MarkdownDescription: "Space ID",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Description:         "Role name",
				MarkdownDescription: "Role name",
			},
		},
	}
}

type RolesModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedOn   types.String `tfsdk:"created_on"`
	UpdatedOn   types.String `tfsdk:"updated_on"`
}

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
