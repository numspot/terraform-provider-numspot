package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_dhcp_options"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func DhcpOptionsFromHttpToTf(ctx context.Context, http *api.DhcpOptionsSet) (*resource_dhcp_options.DhcpOptionsModel, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var domainNameServersTf, logServersTf, ntpServersTf types.List

	if http.DomainNameServers != nil {
		domainNameServersTf, diagnostics = utils.StringListToTfListValue(ctx, *http.DomainNameServers)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
	}

	if http.LogServers != nil {
		logServersTf, diagnostics = utils.StringListToTfListValue(ctx, *http.LogServers)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
	}

	if http.NtpServers != nil {
		ntpServersTf, diagnostics = utils.StringListToTfListValue(ctx, *http.LogServers)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
	}

	return &resource_dhcp_options.DhcpOptionsModel{
		Default:           types.BoolPointerValue(http.Default),
		DomainName:        types.StringPointerValue(http.DomainName),
		Id:                types.StringPointerValue(http.Id),
		DomainNameServers: domainNameServersTf,
		LogServers:        logServersTf,
		NtpServers:        ntpServersTf,
	}, nil
}

func DhcpOptionsFromTfToCreateRequest(ctx context.Context, tf resource_dhcp_options.DhcpOptionsModel) api.CreateDhcpOptionsJSONRequestBody {
	var domainNameServers, logServers, ntpServers []string

	domainNameServers = make([]string, 0, len(tf.DomainNameServers.Elements()))
	tf.DomainNameServers.ElementsAs(ctx, &domainNameServers, false)

	logServers = make([]string, 0, len(tf.LogServers.Elements()))
	tf.LogServers.ElementsAs(ctx, &logServers, false)

	ntpServers = make([]string, 0, len(tf.NtpServers.Elements()))
	tf.NtpServers.ElementsAs(ctx, &ntpServers, false)

	return api.CreateDhcpOptionsJSONRequestBody{
		DomainName:        tf.DomainName.ValueStringPointer(),
		DomainNameServers: &domainNameServers,
		LogServers:        &logServers,
		NtpServers:        &ntpServers,
	}
}
