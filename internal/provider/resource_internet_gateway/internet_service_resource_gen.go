// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_internet_gateway

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func InternetGatewayResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Internet service.",
				MarkdownDescription: "The ID of the Internet service.",
			},
			"vpc_id": schema.StringAttribute{
				Optional:            true,
				Description:         "The ID of the Net attached to the Internet service.",
				MarkdownDescription: "The ID of the Net attached to the Internet service.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the attachment of the Internet service to the Net (always `available`).",
				MarkdownDescription: "The state of the attachment of the Internet service to the Net (always `available`).",
			},
			"tags": tags.TagsSchema(ctx),
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated",
	}
}

type InternetGatewayModel struct {
	Id    types.String `tfsdk:"id"`
	VpcIp types.String `tfsdk:"vpc_id"`
	State types.String `tfsdk:"state"`
	Tags  types.List   `tfsdk:"tags"`
}
