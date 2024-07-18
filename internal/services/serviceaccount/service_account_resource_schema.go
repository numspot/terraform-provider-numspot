package serviceaccount

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ServiceAccountResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
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
			"secret": schema.StringAttribute{
				Computed:            true,
				Description:         "Service account secret.",
				MarkdownDescription: "Service account secret.",
			},
			"space_id": schema.StringAttribute{
				Required:            true,
				Description:         "Space ID",
				MarkdownDescription: "Space ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"global_permissions": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "List of global permission UUIDs attached to this service account.",
				MarkdownDescription: "List of global permission UUIDs attached to this service account.",
			},
			"roles": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "List of role attached to this service account.",
				MarkdownDescription: "List of roles attached to this service account.",
			},
			"service_account_id": schema.StringAttribute{
				Computed:            true,
				Description:         "Service account ID",
				MarkdownDescription: "Service account ID",
			}, // TODO: What if a resource uses the service account id to reference a service account. User should know what id to use between space_id and service_account_id
		},
	}
}

type ServiceAccountModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Secret            types.String `tfsdk:"secret"`
	SpaceId           types.String `tfsdk:"space_id"`
	ServiceAccountId  types.String `tfsdk:"service_account_id"`
	GlobalPermissions types.Set    `tfsdk:"global_permissions"`
	Roles             types.Set    `tfsdk:"roles"`
}
