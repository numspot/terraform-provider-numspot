package serviceaccount

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ServiceAccountDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
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
				}, // MANUALLY EDITED : Removed CustomType block
			},
			"space_id": schema.StringAttribute{
				Required:            true,
				Description:         "Space ID",
				MarkdownDescription: "Space ID",
			},
			"service_account_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Service account IDs",
				MarkdownDescription: "Service account IDs",
			},
		},
	}
}

type ServiceAccountDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
