package securitygroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SecurityGroupFromTfToHttp(tf *SecurityGroupModel) *numspot.SecurityGroup {
	return &numspot.SecurityGroup{
		Id:            tf.Id.ValueStringPointer(),
		Description:   tf.Description.ValueStringPointer(),
		Name:          tf.Name.ValueStringPointer(),
		VpcId:         tf.VpcId.ValueStringPointer(),
		InboundRules:  nil,
		OutboundRules: nil,
	}
}

func InboundRuleFromHttpToTf(ctx context.Context, rules numspot.SecurityGroupRule) InboundRulesValue {
	ipRanges, diags := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	if diags.HasError() {
		return InboundRulesValue{}
	}

	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return InboundRulesValue{}
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return InboundRulesValue{}
		}
	}

	value, diagnostics := NewInboundRulesValue(
		InboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range":         utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":           utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":             types.StringPointerValue(rules.IpProtocol),
			"ip_ranges":               ipRanges,
			"security_groups_members": types.ListNull(SecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":             serviceIds,
		},
	)
	diags.Append(diagnostics...)
	return value
}

func OutboundRuleFromHttpToTf(ctx context.Context, rules numspot.SecurityGroupRule) OutboundRulesValue {
	ipRanges, diags := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	if diags.HasError() {
		return OutboundRulesValue{}
	}

	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return OutboundRulesValue{}
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return OutboundRulesValue{}
		}
	}

	value, diagnostics := NewOutboundRulesValue(
		OutboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range":         utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":           utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":             types.StringPointerValue(rules.IpProtocol),
			"ip_ranges":               ipRanges,
			"security_groups_members": types.ListNull(SecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":             serviceIds,
		},
	)
	diags.Append(diagnostics...)
	return value
}

func SecurityGroupFromHttpToTf(ctx context.Context, http *numspot.SecurityGroup, diags *diag.Diagnostics) *SecurityGroupModel {
	var tagsTf types.List

	if http.InboundRules == nil {
		return nil
	}
	ibd := make([]InboundRulesValue, 0, len(*http.InboundRules))
	for _, e := range *http.InboundRules {
		value := InboundRuleFromHttpToTf(ctx, e)
		if diags.HasError() {
			return nil
		}

		ibd = append(ibd, value)
	}

	if http.OutboundRules == nil {
		return nil
	}
	obd := make([]OutboundRulesValue, 0, len(*http.OutboundRules))
	for _, e := range *http.OutboundRules {
		value := OutboundRuleFromHttpToTf(ctx, e)
		if diags.HasError() {
			return nil
		}

		obd = append(obd, value)
	}

	ibdsTf, diagnostics := types.SetValueFrom(ctx, InboundRulesValue{}.Type(ctx), ibd)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return nil
	}

	obdsTf, diagnostics := types.SetValueFrom(ctx, OutboundRulesValue{}.Type(ctx), obd)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return nil
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	res := SecurityGroupModel{
		Description:   types.StringPointerValue(http.Description),
		Id:            types.StringPointerValue(http.Id),
		Name:          types.StringPointerValue(http.Name),
		VpcId:         types.StringPointerValue(http.VpcId),
		InboundRules:  ibdsTf,
		OutboundRules: obdsTf,
		Tags:          tagsTf,
	}

	return &res
}

func SecurityGroupFromTfToCreateRequest(tf *SecurityGroupModel) numspot.CreateSecurityGroupJSONRequestBody {
	return numspot.CreateSecurityGroupJSONRequestBody{
		Description: tf.Description.ValueString(),
		VpcId:       tf.VpcId.ValueString(),
		Name:        tf.Name.ValueString(),
	}
}

func CreateInboundRulesRequest(ctx context.Context, sgId string, data []InboundRulesValue) numspot.CreateSecurityGroupRuleJSONRequestBody {
	rules := make([]numspot.SecurityGroupRule, 0, len(data))
	for i := range data {
		e := &data[i]
		fpr := int(e.FromPortRange.ValueInt64())
		tpr := int(e.ToPortRange.ValueInt64())

		tfIpRange := make([]types.String, 0, len(e.IpRanges.Elements()))
		e.IpRanges.ElementsAs(ctx, &tfIpRange, false)
		var ipRanges []string
		for _, ip := range tfIpRange {
			ipRanges = append(ipRanges, ip.ValueString())
		}

		rules = append(rules, numspot.SecurityGroupRule{
			FromPortRange: &fpr,
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			IpRanges:      &ipRanges,
			ToPortRange:   &tpr,
		})
	}
	tt := 22
	inboundRulesCreationBody := numspot.CreateSecurityGroupRuleJSONRequestBody{
		Flow:          "Inbound",
		Rules:         &rules,
		FromPortRange: &tt,
		ToPortRange:   &tt,
	}

	return inboundRulesCreationBody
}

func CreateOutboundRulesRequest(ctx context.Context, sgId string, data []OutboundRulesValue) numspot.CreateSecurityGroupRuleJSONRequestBody {
	rules := make([]numspot.SecurityGroupRule, 0, len(data))
	for i := range data {
		e := &data[i]

		fpr := int(e.FromPortRange.ValueInt64())
		tpr := int(e.ToPortRange.ValueInt64())

		tfIpRange := make([]types.String, 0, len(e.IpRanges.Elements()))
		e.IpRanges.ElementsAs(ctx, &tfIpRange, false)
		var ipRanges []string
		for _, ip := range tfIpRange {
			ipRanges = append(ipRanges, ip.ValueString())
		}

		rules = append(rules, numspot.SecurityGroupRule{
			FromPortRange: &fpr,
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			IpRanges:      &ipRanges,
			ToPortRange:   &tpr,
		})
	}

	outboundRulesCreationBody := numspot.CreateSecurityGroupRuleJSONRequestBody{
		Flow:  "Outbound",
		Rules: &rules,
	}

	return outboundRulesCreationBody
}

func SecurityGroupsFromTfToAPIReadParams(ctx context.Context, tf SecurityGroupsDataSourceModel, diags *diag.Diagnostics) numspot.ReadSecurityGroupsParams {
	return numspot.ReadSecurityGroupsParams{
		Descriptions:                 utils.TfStringListToStringPtrList(ctx, tf.Descriptions, diags),
		InboundRuleFromPortRanges:    utils.TFInt64ListToIntListPointer(ctx, tf.InboundRuleFromPortRanges, diags),
		InboundRuleProtocols:         utils.TfStringListToStringPtrList(ctx, tf.InboundRuleProtocols, diags),
		InboundRuleIpRanges:          utils.TfStringListToStringPtrList(ctx, tf.InboundRuleIpRanges, diags),
		InboundRuleSecurityGroupIds:  utils.TfStringListToStringPtrList(ctx, tf.InboundRuleSecurityGroupIds, diags),
		InboundRuleToPortRanges:      utils.TFInt64ListToIntListPointer(ctx, tf.InboundRuleToPortRanges, diags),
		SecurityGroupIds:             utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:           utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames, diags),
		OutboundRuleFromPortRanges:   utils.TFInt64ListToIntListPointer(ctx, tf.OutboundRuleFromPortRanges, diags),
		OutboundRuleProtocols:        utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleProtocols, diags),
		OutboundRuleIpRanges:         utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleIpRanges, diags),
		OutboundRuleSecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleSecurityGroupIds, diags),
		OutboundRuleToPortRanges:     utils.TFInt64ListToIntListPointer(ctx, tf.OutboundRuleToPortRanges, diags),
		VpcIds:                       utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		TagKeys:                      utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:                    utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:                         utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
	}
}

func SecurityGroupsFromHttpToTfDatasource(ctx context.Context, http *numspot.SecurityGroup, diags *diag.Diagnostics) *SecurityGroupModel {
	var (
		inboundRules  = types.SetNull(InboundRulesValue{}.Type(ctx))
		outboundRules = types.SetNull(OutboundRulesValue{}.Type(ctx))
		tagsList      types.List
	)

	if http.InboundRules != nil {
		inboundRules = utils.GenericSetToTfSetValue(
			ctx,
			inboundRuleFromHttpToTfDatasource,
			*http.InboundRules,
			diags,
		)
		if diags.HasError() {
			return nil
		}
	}

	if http.OutboundRules != nil {
		outboundRules = utils.GenericSetToTfSetValue(
			ctx,
			outboundRuleFromHttpToTfDatasource,
			*http.OutboundRules,
			diags,
		)
		if diags.HasError() {
			return nil
		}
	}

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &SecurityGroupModel{
		Id:            types.StringPointerValue(http.Id),
		Tags:          tagsList,
		Description:   types.StringPointerValue(http.Description),
		InboundRules:  inboundRules,
		Name:          types.StringPointerValue(http.Name),
		OutboundRules: outboundRules,
		VpcId:         types.StringPointerValue(http.VpcId),
	}
}

func inboundRuleFromHttpToTfDatasource(ctx context.Context, rules numspot.SecurityGroupRule, diags *diag.Diagnostics) InboundRulesValue {
	ipRanges, diagnostics := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	diags.Append(diagnostics...)
	serviceIds, diagnostics := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	diags.Append(diagnostics...)
	if rules.ServiceIds == nil {
		serviceIds, diagnostics = types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(diagnostics...)
	}

	value, diagnostics := NewInboundRulesValue(
		InboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range":         utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":           utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":             types.StringPointerValue(rules.IpProtocol),
			"ip_ranges":               ipRanges,
			"security_groups_members": types.ListNull(SecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":             serviceIds,
		},
	)
	diags.Append(diagnostics...)
	return value
}

func outboundRuleFromHttpToTfDatasource(ctx context.Context, rules numspot.SecurityGroupRule, diags *diag.Diagnostics) OutboundRulesValue {
	ipRanges, diagnostics := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	diags.Append(diagnostics...)

	serviceIds, diagnostics := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	diags.Append(diagnostics...)

	if rules.ServiceIds == nil {
		serviceIds, diagnostics = types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(diagnostics...)
	}

	value, diagnostics := NewOutboundRulesValue(
		OutboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range":         utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":           utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":             types.StringPointerValue(rules.IpProtocol),
			"ip_ranges":               ipRanges,
			"security_groups_members": types.ListNull(SecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":             serviceIds,
		},
	)
	diags.Append(diagnostics...)
	return value
}
