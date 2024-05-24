package datasource_service_account

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ServiceAccountDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_accounts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "Service account ID",
							MarkdownDescription: "Service account ID",
						},
						"name": schema.StringAttribute{
							Required:            true,
							Description:         "Service Account name",
							MarkdownDescription: "Service Account name",
						},
					},
				},
			},
			"space_id": schema.StringAttribute{
				Required:            true,
				Description:         "Space ID",
				MarkdownDescription: "Space ID",
			},
			"service_account_name": schema.StringAttribute{
				Optional:            true,
				Description:         "Service account name",
				MarkdownDescription: "Service account name",
			},
		},
	}
}

type ServiceAccountModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
