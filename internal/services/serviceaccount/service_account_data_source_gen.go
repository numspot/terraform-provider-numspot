package serviceaccount

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
			"service_account_name": schema.StringAttribute{
				Optional:            true,
				Description:         "Service account name",
				MarkdownDescription: "Service account name",
			},
		},
	}
}

// MANUALLY EDITED : Model declaration removed

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
