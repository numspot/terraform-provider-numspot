package dhcpoptions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

//func DhcpOptionsFromHttpToTf(ctx context.Context, http *numspot.DhcpOptionsSet) (*DhcpOptionsModel, diag.Diagnostics) {
//	var diagnostics diag.Diagnostics
//	var domainNameServersTf, logServersTf, ntpServersTf, tagsTf types.List
//
//	if http.DomainNameServers != nil {
//		domainNameServersTf, diagnostics = utils.StringListToTfListValue(ctx, *http.DomainNameServers)
//		if diagnostics.HasError() {
//			return nil, diagnostics
//		}
//	}
//
//	if http.LogServers != nil {
//		logServersTf, diagnostics = utils.StringListToTfListValue(ctx, *http.LogServers)
//		if diagnostics.HasError() {
//			return nil, diagnostics
//		}
//	}
//
//	if http.NtpServers != nil {
//		ntpServersTf, diagnostics = utils.StringListToTfListValue(ctx, *http.LogServers)
//		if diagnostics.HasError() {
//			return nil, diagnostics
//		}
//	}
//
//	if http.Tags != nil {
//		tagsTf, diagnostics = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
//		if diagnostics.HasError() {
//			return nil, diagnostics
//		}
//	}
//
//	return &DhcpOptionsModel{
//		Default:           types.BoolPointerValue(http.Default),
//		DomainName:        types.StringPointerValue(http.DomainName),
//		Id:                types.StringPointerValue(http.Id),
//		DomainNameServers: domainNameServersTf,
//		LogServers:        logServersTf,
//		NtpServers:        ntpServersTf,
//		Tags:              tagsTf,
//	}, nil
//}

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
