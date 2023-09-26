// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_internet_service

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func InternetServiceResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Internet service.",
				MarkdownDescription: "The ID of the Internet service.",
			},
			"net_id": schema.StringAttribute{
				Optional:            true,
				Description:         "The ID of the Net attached to the Internet service.",
				MarkdownDescription: "The ID of the Net attached to the Internet service.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the attachment of the Internet service to the Net (always `available`).",
				MarkdownDescription: "The state of the attachment of the Internet service to the Net (always `available`).",
			},
		},
	}
}

type InternetServiceModel struct {
	Id    types.String `tfsdk:"id"`
	NetId types.String `tfsdk:"net_id"`
	State types.String `tfsdk:"state"`
}
