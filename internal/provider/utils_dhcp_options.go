package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_dhcp_options"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_dhcp_options"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func DhcpOptionsFromHttpToTf(ctx context.Context, http *numspot.DhcpOptionsSet) (*resource_dhcp_options.DhcpOptionsModel, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var domainNameServersTf, logServersTf, ntpServersTf, tagsTf types.List

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

	if http.Tags != nil {
		tagsTf, diagnostics = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
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
		Tags:              tagsTf,
	}, nil
}

func DhcpOptionsFromTfToCreateRequest(ctx context.Context, tf resource_dhcp_options.DhcpOptionsModel) numspot.CreateDhcpOptionsJSONRequestBody {
	var domainNameServers, logServers, ntpServers []string

	domainNameServers = make([]string, 0, len(tf.DomainNameServers.Elements()))
	tf.DomainNameServers.ElementsAs(ctx, &domainNameServers, false)

	logServers = make([]string, 0, len(tf.LogServers.Elements()))
	tf.LogServers.ElementsAs(ctx, &logServers, false)

	ntpServers = make([]string, 0, len(tf.NtpServers.Elements()))
	tf.NtpServers.ElementsAs(ctx, &ntpServers, false)

	return numspot.CreateDhcpOptionsJSONRequestBody{
		DomainName:        tf.DomainName.ValueStringPointer(),
		DomainNameServers: &domainNameServers,
		LogServers:        &logServers,
		NtpServers:        &ntpServers,
	}
}

func DhcpOptionsFromTfToAPIReadParams(ctx context.Context, tf DHCPOptionsDataSourceModel) numspot.ReadDhcpOptionsParams {
	ids := utils.TfStringListToStringPtrList(ctx, tf.IDs)
	domainNames := utils.TfStringListToStringPtrList(ctx, tf.DomainNames)
	dnsServers := utils.TfStringListToStringPtrList(ctx, tf.DomainNameServers)
	logServers := utils.TfStringListToStringPtrList(ctx, tf.LogServers)
	ntpServers := utils.TfStringListToStringPtrList(ctx, tf.NTPServers)
	tagKeys := utils.TfStringListToStringPtrList(ctx, tf.TagKeys)
	tagValues := utils.TfStringListToStringPtrList(ctx, tf.TagValues)
	tags := utils.TfStringListToStringPtrList(ctx, tf.Tags)

	return numspot.ReadDhcpOptionsParams{
		Default:           tf.Default.ValueBoolPointer(),
		DomainNameServers: dnsServers,
		DomainNames:       domainNames,
		LogServers:        logServers,
		NtpServers:        ntpServers,
		TagKeys:           tagKeys,
		TagValues:         tagValues,
		Tags:              tags,
		Ids:               ids,
	}
}

func DHCPOptionsFromHttpToTfDatasource(ctx context.Context, http *numspot.DhcpOptionsSet) (*datasource_dhcp_options.DhcpOptionsModel, diag.Diagnostics) {
	var tagsList types.List
	dnsServers, diags := utils.FromStringListPointerToTfStringList(ctx, http.DomainNameServers)
	if diags.HasError() {
		return nil, diags
	}
	logServers, diags := utils.FromStringListPointerToTfStringList(ctx, http.LogServers)
	if diags.HasError() {
		return nil, diags
	}
	ntpServers, diags := utils.FromStringListPointerToTfStringList(ctx, http.NtpServers)
	if diags.HasError() {
		return nil, diags
	}
	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}
	return &datasource_dhcp_options.DhcpOptionsModel{
		Default:           types.BoolPointerValue(http.Default),
		DomainName:        types.StringPointerValue(http.DomainName),
		DomainNameServers: dnsServers,
		Id:                types.StringPointerValue(http.Id),
		LogServers:        logServers,
		NtpServers:        ntpServers,
		Tags:              tagsList,
	}, nil
}
