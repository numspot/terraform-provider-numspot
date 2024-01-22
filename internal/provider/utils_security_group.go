package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_security_group"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SecurityGroupFromTfToHttp(tf resource_security_group.SecurityGroupModel) *api.SecurityGroupSchema {
	return &api.SecurityGroupSchema{
		Id:            tf.Id.ValueStringPointer(),
		AccountId:     tf.AccountId.ValueStringPointer(),
		Description:   tf.Description.ValueStringPointer(),
		Name:          tf.Name.ValueStringPointer(),
		NetId:         tf.NetId.ValueStringPointer(),
		InboundRules:  nil,
		OutboundRules: nil,
	}
}

func InboundRuleFromHttpToTf(ctx context.Context, rules api.SecurityGroupRuleSchema) (resource_security_group.InboundRulesValue, diag.Diagnostics) {
	ipRanges, diags := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	if diags.HasError() {
		return resource_security_group.InboundRulesValue{}, diags
	}

	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return resource_security_group.InboundRulesValue{}, diags
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return resource_security_group.InboundRulesValue{}, diags
		}
	}

	return resource_security_group.NewInboundRulesValue(
		resource_security_group.InboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range":         utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":           utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":             types.StringPointerValue(rules.IpProtocol),
			"ip_ranges":               ipRanges,
			"security_groups_members": types.ListNull(resource_security_group.SecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":             serviceIds,
		},
	)
}

func OutboundRuleFromHttpToTf(ctx context.Context, rules api.SecurityGroupRuleSchema) (resource_security_group.OutboundRulesValue, diag.Diagnostics) {
	ipRanges, diags := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	if diags.HasError() {
		return resource_security_group.OutboundRulesValue{}, diags
	}

	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return resource_security_group.OutboundRulesValue{}, diags
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return resource_security_group.OutboundRulesValue{}, diags
		}
	}

	return resource_security_group.NewOutboundRulesValue(
		resource_security_group.OutboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range":         utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":           utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":             types.StringPointerValue(rules.IpProtocol),
			"ip_ranges":               ipRanges,
			"security_groups_members": types.ListNull(resource_security_group.SecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":             serviceIds,
		},
	)
}

func SecurityGroupFromHttpToTf(ctx context.Context, http *api.SecurityGroupSchema) (*resource_security_group.SecurityGroupModel, diag.Diagnostics) {
	ibds := make([]resource_security_group.InboundRulesValue, 0, len(*http.InboundRules))
	for _, e := range *http.InboundRules {
		value, diag := InboundRuleFromHttpToTf(ctx, e)
		if diag.HasError() {
			return nil, diag
		}

		ibds = append(ibds, value)
	}

	obds := make([]resource_security_group.OutboundRulesValue, 0, len(*http.OutboundRules))
	for _, e := range *http.OutboundRules {
		value, diag := OutboundRuleFromHttpToTf(ctx, e)
		if diag.HasError() {
			return nil, diag
		}

		obds = append(obds, value)
	}

	ibdsTf, diag := types.ListValueFrom(ctx, resource_security_group.InboundRulesValue{}.Type(ctx), ibds)
	if diag.HasError() {
		return nil, diag
	}

	obdsTf, diag := types.ListValueFrom(ctx, resource_security_group.OutboundRulesValue{}.Type(ctx), obds)
	if diag.HasError() {
		return nil, diag
	}

	res := resource_security_group.SecurityGroupModel{
		AccountId:     types.StringPointerValue(http.AccountId),
		Description:   types.StringPointerValue(http.Description),
		Id:            types.StringPointerValue(http.Id),
		Name:          types.StringPointerValue(http.Name),
		NetId:         types.StringPointerValue(http.NetId),
		InboundRules:  ibdsTf,
		OutboundRules: obdsTf,
	}

	return &res, nil
}

func SecurityGroupFromTfToCreateRequest(tf resource_security_group.SecurityGroupModel) api.CreateSecurityGroupJSONRequestBody {
	return api.CreateSecurityGroupJSONRequestBody{
		Description: tf.Description.ValueString(),
		NetId:       tf.NetId.ValueStringPointer(),
		Name:        tf.Name.ValueStringPointer(),
	}
}

func CreateInboundRulesRequest(ctx context.Context, sgId string, data []resource_security_group.InboundRulesValue) api.CreateSecurityGroupRuleJSONRequestBody {
	rules := make([]api.SecurityGroupRuleSchema, 0, len(data))
	for _, e := range data {
		fpr := int(e.FromPortRange.ValueInt64())
		tpr := int(e.ToPortRange.ValueInt64())

		tfIpRange := make([]types.String, 0, len(e.IpRanges.Elements()))
		e.IpRanges.ElementsAs(ctx, &tfIpRange, false)
		var ipRanges []string
		for _, ip := range tfIpRange {
			ipRanges = append(ipRanges, ip.ValueString())
		}

		rules = append(rules, api.SecurityGroupRuleSchema{
			FromPortRange: &fpr,
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			IpRanges:      &ipRanges,
			ToPortRange:   &tpr,
		})
	}
	tt := 22
	inboundRulesCreationBody := api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:            "Inbound",
		SecurityGroupId: sgId,
		Rules:           &rules,
		FromPortRange:   &tt,
		ToPortRange:     &tt,
	}

	return inboundRulesCreationBody
}

func CreateOutboundRulesRequest(ctx context.Context, sgId string, data []resource_security_group.OutboundRulesValue) api.CreateSecurityGroupRuleJSONRequestBody {
	rules := make([]api.SecurityGroupRuleSchema, 0, len(data))
	for _, e := range data {
		fpr := int(e.FromPortRange.ValueInt64())
		tpr := int(e.ToPortRange.ValueInt64())

		tfIpRange := make([]types.String, 0, len(e.IpRanges.Elements()))
		e.IpRanges.ElementsAs(ctx, &tfIpRange, false)
		var ipRanges []string
		for _, ip := range tfIpRange {
			ipRanges = append(ipRanges, ip.ValueString())
		}

		rules = append(rules, api.SecurityGroupRuleSchema{
			FromPortRange: &fpr,
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			IpRanges:      &ipRanges,
			ToPortRange:   &tpr,
		})
	}

	outboundRulesCreationBody := api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:            "Outbound",
		SecurityGroupId: sgId,
		Rules:           &rules,
	}

	return outboundRulesCreationBody
}
