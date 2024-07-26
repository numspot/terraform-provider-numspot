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

func InboundRuleFromHttpToTf(ctx context.Context, rules numspot.SecurityGroupRule) (InboundRulesValue, diag.Diagnostics) {
	ipRanges, diags := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	if diags.HasError() {
		return InboundRulesValue{}, diags
	}

	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return InboundRulesValue{}, diags
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return InboundRulesValue{}, diags
		}
	}

	return NewInboundRulesValue(
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
}

func OutboundRuleFromHttpToTf(ctx context.Context, rules numspot.SecurityGroupRule) (OutboundRulesValue, diag.Diagnostics) {
	ipRanges, diags := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	if diags.HasError() {
		return OutboundRulesValue{}, diags
	}

	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return OutboundRulesValue{}, diags
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return OutboundRulesValue{}, diags
		}
	}

	return NewOutboundRulesValue(
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
}

func SecurityGroupFromHttpToTf(ctx context.Context, http *numspot.SecurityGroup) (*SecurityGroupModel, diag.Diagnostics) {
	var (
		tagsTf types.List
		diags  diag.Diagnostics
	)

	if http.InboundRules == nil {
		return nil, diags
	}
	ibd := make([]InboundRulesValue, 0, len(*http.InboundRules))
	for _, e := range *http.InboundRules {
		value, diags := InboundRuleFromHttpToTf(ctx, e)
		if diags.HasError() {
			return nil, diags
		}

		ibd = append(ibd, value)
	}

	if http.OutboundRules == nil {
		return nil, diags
	}
	obd := make([]OutboundRulesValue, 0, len(*http.OutboundRules))
	for _, e := range *http.OutboundRules {
		value, diags := OutboundRuleFromHttpToTf(ctx, e)
		if diags.HasError() {
			return nil, diags
		}

		obd = append(obd, value)
	}

	ibdsTf, diags := types.SetValueFrom(ctx, InboundRulesValue{}.Type(ctx), ibd)
	if diags.HasError() {
		return nil, diags
	}

	obdsTf, diags := types.SetValueFrom(ctx, OutboundRulesValue{}.Type(ctx), obd)
	if diags.HasError() {
		return nil, diags
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
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

	return &res, diags
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

func SecurityGroupsFromTfToAPIReadParams(ctx context.Context, tf SecurityGroupsDataSourceModel) numspot.ReadSecurityGroupsParams {
	return numspot.ReadSecurityGroupsParams{
		Descriptions:                 utils.TfStringListToStringPtrList(ctx, tf.Descriptions),
		InboundRuleFromPortRanges:    utils.TFInt64ListToIntListPointer(ctx, tf.InboundRuleFromPortRanges),
		InboundRuleProtocols:         utils.TfStringListToStringPtrList(ctx, tf.InboundRuleProtocols),
		InboundRuleIpRanges:          utils.TfStringListToStringPtrList(ctx, tf.InboundRuleIpRanges),
		InboundRuleSecurityGroupIds:  utils.TfStringListToStringPtrList(ctx, tf.InboundRuleSecurityGroupIds),
		InboundRuleToPortRanges:      utils.TFInt64ListToIntListPointer(ctx, tf.InboundRuleToPortRanges),
		SecurityGroupIds:             utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		SecurityGroupNames:           utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames),
		OutboundRuleFromPortRanges:   utils.TFInt64ListToIntListPointer(ctx, tf.OutboundRuleFromPortRanges),
		OutboundRuleProtocols:        utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleProtocols),
		OutboundRuleIpRanges:         utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleIpRanges),
		OutboundRuleSecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleSecurityGroupIds),
		OutboundRuleToPortRanges:     utils.TFInt64ListToIntListPointer(ctx, tf.OutboundRuleToPortRanges),
		VpcIds:                       utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		TagKeys:                      utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:                    utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:                         utils.TfStringListToStringPtrList(ctx, tf.Tags),
	}
}

func SecurityGroupsFromHttpToTfDatasource(ctx context.Context, http *numspot.SecurityGroup) (*SecurityGroupModel, diag.Diagnostics) {
	var (
		inboundRules  = types.SetNull(InboundRulesValue{}.Type(ctx))
		outboundRules = types.SetNull(OutboundRulesValue{}.Type(ctx))
		diags         diag.Diagnostics
		tagsList      types.List
	)

	if http.InboundRules != nil {
		inboundRules, diags = utils.GenericSetToTfSetValue(
			ctx,
			InboundRulesValue{},
			inboundRuleFromHttpToTfDatasource,
			*http.InboundRules,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.OutboundRules != nil {
		outboundRules, diags = utils.GenericSetToTfSetValue(
			ctx,
			OutboundRulesValue{},
			outboundRuleFromHttpToTfDatasource,
			*http.OutboundRules,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
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
	}, nil
}

func inboundRuleFromHttpToTfDatasource(ctx context.Context, rules numspot.SecurityGroupRule) (InboundRulesValue, diag.Diagnostics) {
	ipRanges, diags := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	if diags.HasError() {
		return InboundRulesValue{}, diags
	}

	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return InboundRulesValue{}, diags
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return InboundRulesValue{}, diags
		}
	}

	return NewInboundRulesValue(
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
}

func outboundRuleFromHttpToTfDatasource(ctx context.Context, rules numspot.SecurityGroupRule) (OutboundRulesValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	ipRanges, diags := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	if diags.HasError() {
		return OutboundRulesValue{}, diags
	}

	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return OutboundRulesValue{}, diags
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return OutboundRulesValue{}, diags
		}
	}

	return NewOutboundRulesValue(
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
}
