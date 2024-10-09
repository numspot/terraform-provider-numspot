package dhcpoptions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func DhcpOptionsFromHttpToTf(ctx context.Context, http *numspot.DhcpOptionsSet, diags *diag.Diagnostics) *DhcpOptionsModel {
	var domainNameServersTf, logServersTf, ntpServersTf, tagsTf types.List

	if http.DomainNameServers != nil {
		domainNameServersTf = utils.StringListToTfListValue(ctx, *http.DomainNameServers, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.LogServers != nil {
		logServersTf = utils.StringListToTfListValue(ctx, *http.LogServers, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.NtpServers != nil {
		ntpServersTf = utils.StringListToTfListValue(ctx, *http.LogServers, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &DhcpOptionsModel{
		Default:           types.BoolPointerValue(http.Default),
		DomainName:        types.StringPointerValue(http.DomainName),
		Id:                types.StringPointerValue(http.Id),
		DomainNameServers: domainNameServersTf,
		LogServers:        logServersTf,
		NtpServers:        ntpServersTf,
		Tags:              tagsTf,
	}
}

func DhcpOptionsFromTfToCreateRequest(ctx context.Context, tf DhcpOptionsModel) numspot.CreateDhcpOptionsJSONRequestBody {
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

func DhcpOptionsFromTfToAPIReadParams(ctx context.Context, tf DHCPOptionsDataSourceModel, diags *diag.Diagnostics) numspot.ReadDhcpOptionsParams {
	ids := utils.TfStringListToStringPtrList(ctx, tf.IDs, diags)
	domainNames := utils.TfStringListToStringPtrList(ctx, tf.DomainNames, diags)
	dnsServers := utils.TfStringListToStringPtrList(ctx, tf.DomainNameServers, diags)
	logServers := utils.TfStringListToStringPtrList(ctx, tf.LogServers, diags)
	ntpServers := utils.TfStringListToStringPtrList(ctx, tf.NTPServers, diags)
	tagKeys := utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags)
	tagValues := utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags)
	tags := utils.TfStringListToStringPtrList(ctx, tf.Tags, diags)

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

func DHCPOptionsFromHttpToTfDatasource(ctx context.Context, http *numspot.DhcpOptionsSet, diags *diag.Diagnostics) *DhcpOptionsModel {
	return &DhcpOptionsModel{
		Default:           types.BoolPointerValue(http.Default),
		DomainName:        types.StringPointerValue(http.DomainName),
		DomainNameServers: utils.FromStringListPointerToTfStringList(ctx, http.DomainNameServers, diags),
		Id:                types.StringPointerValue(http.Id),
		LogServers:        utils.FromStringListPointerToTfStringList(ctx, http.LogServers, diags),
		NtpServers:        utils.FromStringListPointerToTfStringList(ctx, http.NtpServers, diags),
		Tags:              utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags),
	}
}
