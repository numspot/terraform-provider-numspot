// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package dhcpoptions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
)

func DhcpOptionsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"default": schema.BoolAttribute{
							Computed:            true,
							Description:         "If true, the DHCP options set is a default one. If false, it is not.",
							MarkdownDescription: "If true, the DHCP options set is a default one. If false, it is not.",
						},
						"domain_name": schema.StringAttribute{
							Computed:            true,
							Description:         "The domain name.",
							MarkdownDescription: "The domain name.",
						},
						"domain_name_servers": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "One or more IPs for the domain name servers.",
							MarkdownDescription: "One or more IPs for the domain name servers.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the DHCP options set.",
							MarkdownDescription: "The ID of the DHCP options set.",
						},
						"log_servers": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "One or more IPs for the log servers.",
							MarkdownDescription: "One or more IPs for the log servers.",
						},
						"ntp_servers": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "One or more IPs for the NTP servers.",
							MarkdownDescription: "One or more IPs for the NTP servers.",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more DHCP options sets.",
				MarkdownDescription: "Information about one or more DHCP options sets.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the DHCP options sets.",
				MarkdownDescription: "The IDs of the DHCP options sets.",
			},
			"default": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "If true, lists all default DHCP options set. If false, lists all non-default DHCP options set.",
				MarkdownDescription: "If true, lists all default DHCP options set. If false, lists all non-default DHCP options set.",
			},
			"domain_name_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IPs of the domain name servers used for the DHCP options sets.",
				MarkdownDescription: "The IPs of the domain name servers used for the DHCP options sets.",
			},
			"domain_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The domain names used for the DHCP options sets.",
				MarkdownDescription: "The domain names used for the DHCP options sets.",
			},

			"log_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IPs of the log servers used for the DHCP options sets.",
				MarkdownDescription: "The IPs of the log servers used for the DHCP options sets.",
			},
			"ntp_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IPs of the Network Time Protocol (NTP) servers used for the DHCP options sets.",
				MarkdownDescription: "The IPs of the Network Time Protocol (NTP) servers used for the DHCP options sets.",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the DHCP options sets.",
				MarkdownDescription: "The keys of the tags associated with the DHCP options sets.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the DHCP options sets.",
				MarkdownDescription: "The values of the tags associated with the DHCP options sets.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the DHCP options sets, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the DHCP options sets, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
			// MANUALLY EDITED : SpaceId Removed
		},
	}
}

type DHCPOptionsDataSourceModel struct {
	Items             []DhcpOptionsModel `tfsdk:"items"`
	IDs               types.List         `tfsdk:"ids"`
	Default           types.Bool         `tfsdk:"default"`
	DomainNameServers types.List         `tfsdk:"domain_name_servers"`
	DomainNames       types.List         `tfsdk:"domain_names"`
	LogServers        types.List         `tfsdk:"log_servers"`
	NTPServers        types.List         `tfsdk:"ntp_servers"`
	TagKeys           types.List         `tfsdk:"tag_keys"`
	TagValues         types.List         `tfsdk:"tag_values"`
	Tags              types.List         `tfsdk:"tags"`
}

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
