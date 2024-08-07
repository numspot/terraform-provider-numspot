// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package internetgateway

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func InternetGatewayResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Internet gateway.",
				MarkdownDescription: "The ID of the Internet gateway.",
			},
			"vpc_id": schema.StringAttribute{
				Optional:            true,
				Description:         "The ID of the Vpc attached to the Internet gateway.",
				MarkdownDescription: "The ID of the Vpc attached to the Internet gateway.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the attachment of the Internet gateway to the Vpc (always `available`).",
				MarkdownDescription: "The state of the attachment of the Internet gateway to the Vpc (always `available`).",
			},
			"tags": tags.TagsSchema(ctx),
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated",
	}
}

type InternetGatewayModel struct {
	Id    types.String `tfsdk:"id"`
	VpcId types.String `tfsdk:"vpc_id"`
	State types.String `tfsdk:"state"`
	Tags  types.List   `tfsdk:"tags"`
}
