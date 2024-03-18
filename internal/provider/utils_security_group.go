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

func SecurityGroupFromTfToHttp(tf *resource_security_group.SecurityGroupModel) *api.SecurityGroup {
	return &api.SecurityGroup{
		Id:            tf.Id.ValueStringPointer(),
		AccountId:     tf.AccountId.ValueStringPointer(),
		Description:   tf.Description.ValueStringPointer(),
		Name:          tf.Name.ValueStringPointer(),
		VpcId:         tf.NetId.ValueStringPointer(),
		InboundRules:  nil,
		OutboundRules: nil,
	}
}

func InboundRuleFromHttpToTf(ctx context.Context, rules api.SecurityGroupRule) (resource_security_group.InboundRulesValue, diag.Diagnostics) {
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

func OutboundRuleFromHttpToTf(ctx context.Context, rules api.SecurityGroupRule) (resource_security_group.OutboundRulesValue, diag.Diagnostics) {
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

func SecurityGroupFromHttpToTf(ctx context.Context, model resource_security_group.SecurityGroupModel, http *api.SecurityGroup) (*resource_security_group.SecurityGroupModel, diag.Diagnostics) {
	ibd := make([]resource_security_group.InboundRulesValue, 0, len(*http.InboundRules))
	for _, e := range *http.InboundRules {
		value, diagnostics := InboundRuleFromHttpToTf(ctx, e)
		if diagnostics.HasError() {
			return nil, diagnostics
		}

		ibd = append(ibd, value)
	}

	obd := make([]resource_security_group.OutboundRulesValue, 0, len(*http.OutboundRules))
	for _, e := range *http.OutboundRules {
		value, diagnostics := OutboundRuleFromHttpToTf(ctx, e)
		if diagnostics.HasError() {
			return nil, diagnostics
		}

		obd = append(obd, value)
	}

	// Reordering rules, to match state because osc is reordering security group rules
	if len(model.InboundRules.Elements()) > 0 {
		modelIbd := make([]resource_security_group.InboundRulesValue, 0, len(model.InboundRules.Elements()))
		if diagnostics := model.InboundRules.ElementsAs(ctx, &modelIbd, false); diagnostics.HasError() {
			return nil, diagnostics
		}

		m := true
		for m {
			m = false

			posA := -1
			posB := -1

			for i := range ibd {
				eA := &ibd[i]
				for j := range modelIbd {
					eB := &modelIbd[j]
					if eA.FromPortRange.Equal(eB.FromPortRange) &&
						eA.ToPortRange.Equal(eB.ToPortRange) &&
						eA.IpProtocol.Equal(eB.IpProtocol) &&
						eA.IpRanges.Equal(eB.IpRanges) {
						posA = i
						posB = j
					}
				}
			}

			if posA != -1 && posA != posB {
				ibd[posA], ibd[posB] = ibd[posB], ibd[posA]
				m = true
			}
		}
	}

	if len(model.OutboundRules.Elements()) > 0 {
		modelObd := make([]resource_security_group.OutboundRulesValue, 0, len(model.OutboundRules.Elements()))
		if diagnostics := model.OutboundRules.ElementsAs(ctx, &modelObd, false); diagnostics.HasError() {
			return nil, diagnostics
		}

		m := true
		for m {
			m = false

			posA := -1
			posB := -1

			for i := range obd {
				eA := &ibd[i]
				for j := range modelObd {
					eB := &modelObd[j]
					if eA.FromPortRange.Equal(eB.FromPortRange) &&
						eA.ToPortRange.Equal(eB.ToPortRange) &&
						eA.IpProtocol.Equal(eB.IpProtocol) &&
						eA.IpRanges.Equal(eB.IpRanges) {
						posA = i
						posB = j
					}
				}
			}

			if posA != -1 && posA != posB {
				obd[posA], obd[posB] = obd[posB], obd[posA]
				m = true
			}
		}
	}

	ibdsTf, diagnostics := types.ListValueFrom(ctx, resource_security_group.InboundRulesValue{}.Type(ctx), ibd)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	obdsTf, diagnostics := types.ListValueFrom(ctx, resource_security_group.OutboundRulesValue{}.Type(ctx), obd)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	res := resource_security_group.SecurityGroupModel{
		AccountId:     types.StringPointerValue(http.AccountId),
		Description:   types.StringPointerValue(http.Description),
		Id:            types.StringPointerValue(http.Id),
		Name:          types.StringPointerValue(http.Name),
		NetId:         types.StringPointerValue(http.VpcId),
		InboundRules:  ibdsTf,
		OutboundRules: obdsTf,
	}

	return &res, nil
}

func SecurityGroupFromTfToCreateRequest(tf *resource_security_group.SecurityGroupModel) api.CreateSecurityGroupJSONRequestBody {
	return api.CreateSecurityGroupJSONRequestBody{
		Description: tf.Description.ValueString(),
		VpcId:       tf.NetId.ValueStringPointer(),
		Name:        tf.Name.ValueString(),
	}
}

func CreateInboundRulesRequest(ctx context.Context, sgId string, data []resource_security_group.InboundRulesValue) api.CreateSecurityGroupRuleJSONRequestBody {
	rules := make([]api.SecurityGroupRule, 0, len(data))
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

		rules = append(rules, api.SecurityGroupRule{
			FromPortRange: &fpr,
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			IpRanges:      &ipRanges,
			ToPortRange:   &tpr,
		})
	}
	tt := 22
	inboundRulesCreationBody := api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:          "Inbound",
		Rules:         &rules,
		FromPortRange: &tt,
		ToPortRange:   &tt,
	}

	return inboundRulesCreationBody
}

func CreateOutboundRulesRequest(ctx context.Context, sgId string, data []resource_security_group.OutboundRulesValue) api.CreateSecurityGroupRuleJSONRequestBody {
	rules := make([]api.SecurityGroupRule, 0, len(data))
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

		rules = append(rules, api.SecurityGroupRule{
			FromPortRange: &fpr,
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			IpRanges:      &ipRanges,
			ToPortRange:   &tpr,
		})
	}

	outboundRulesCreationBody := api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:  "Outbound",
		Rules: &rules,
	}

	return outboundRulesCreationBody
}
