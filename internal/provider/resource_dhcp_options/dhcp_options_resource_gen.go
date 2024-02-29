// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_dhcp_options

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DhcpOptionsResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default": schema.BoolAttribute{
				Computed:            true,
				Description:         "If true, the DHCP options set is a default one. If false, it is not.",
				MarkdownDescription: "If true, the DHCP options set is a default one. If false, it is not.",
			},
			"domain_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Specify a domain name (for example, `MyCompany.com`). You can specify only one domain name. You must specify at least one of the following parameters: `DomainName`, `DomainNameServers`, `LogServers`, or `NtpServers`.",
				MarkdownDescription: "Specify a domain name (for example, `MyCompany.com`). You can specify only one domain name. You must specify at least one of the following parameters: `DomainName`, `DomainNameServers`, `LogServers`, or `NtpServers`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_name_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IPs of domain name servers. If no IPs are specified, the `OutscaleProvidedDNS` value is set by default. You must specify at least one of the following parameters: `DomainName`, `DomainNameServers`, `LogServers`, or `NtpServers`.",
				MarkdownDescription: "The IPs of domain name servers. If no IPs are specified, the `OutscaleProvidedDNS` value is set by default. You must specify at least one of the following parameters: `DomainName`, `DomainNameServers`, `LogServers`, or `NtpServers`.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the DHCP options set.",
				MarkdownDescription: "The ID of the DHCP options set.",
			},
			"log_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IPs of the log servers. You must specify at least one of the following parameters: `DomainName`, `DomainNameServers`, `LogServers`, or `NtpServers`.",
				MarkdownDescription: "The IPs of the log servers. You must specify at least one of the following parameters: `DomainName`, `DomainNameServers`, `LogServers`, or `NtpServers`.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"ntp_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IPs of the Network Time Protocol (NTP) servers. You must specify at least one of the following parameters: `DomainName`, `DomainNameServers`, `LogServers`, or `NtpServers`.",
				MarkdownDescription: "The IPs of the Network Time Protocol (NTP) servers. You must specify at least one of the following parameters: `DomainName`, `DomainNameServers`, `LogServers`, or `NtpServers`.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

type DhcpOptionsModel struct {
	Default           types.Bool   `tfsdk:"default"`
	DomainName        types.String `tfsdk:"domain_name"`
	DomainNameServers types.List   `tfsdk:"domain_name_servers"`
	Id                types.String `tfsdk:"id"`
	LogServers        types.List   `tfsdk:"log_servers"`
	NtpServers        types.List   `tfsdk:"ntp_servers"`
}
