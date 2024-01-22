package provider

import (
	"context"
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

func InboundRuleFromHttpToTf(rules api.SecurityGroupRuleSchema) resource_security_group.InboundRulesValue {
	fpr := utils.FromIntPtrToTfInt64(rules.FromPortRange)
	tpr := utils.FromIntPtrToTfInt64(rules.ToPortRange)

	ipRange := types.ListNull(types.StringType)
	if rules.IpRanges != nil {
		ipRange = utils.FromStringListToTfStringList(*rules.IpRanges)
	}

	services := types.ListNull(types.StringType)
	if rules.ServiceIds != nil {
		services = utils.FromStringListToTfStringList(*rules.ServiceIds)
	}

	members := types.ListNull(resource_security_group.SecurityGroupsMembersType{})

	return resource_security_group.InboundRulesValue{
		FromPortRange:         fpr,
		ToPortRange:           tpr,
		IpProtocol:            types.StringPointerValue(rules.IpProtocol),
		IpRanges:              ipRange,
		SecurityGroupsMembers: members,
		ServiceIds:            services,
	}
}

func OutboundRuleFromHttpToTf(ctx context.Context, rules api.SecurityGroupRuleSchema) resource_security_group.OutboundRulesValue {
	fpr := utils.FromIntPtrToTfInt64(rules.FromPortRange)
	tpr := utils.FromIntPtrToTfInt64(rules.ToPortRange)

	ipRange := types.ListNull(types.StringType)
	if rules.IpRanges != nil {
		ipRange = utils.FromStringListToTfStringList(*rules.IpRanges)
	}

	services := types.ListNull(types.StringType)
	if rules.ServiceIds != nil {
		services = utils.FromStringListToTfStringList(*rules.ServiceIds)
	}

	// members := types.ListNull(resource_security_group.SecurityGroupsMembersType{})
	return resource_security_group.OutboundRulesValue{
		FromPortRange: fpr,
		ToPortRange:   tpr,
		IpProtocol:    types.StringPointerValue(rules.IpProtocol),
		IpRanges:      ipRange,
		ServiceIds:    services,
	}
}

func SecurityGroupFromHttpToTf(ctx context.Context, http *api.SecurityGroupSchema) (*resource_security_group.SecurityGroupModel, diag.Diagnostics) {
	ibds := make([]resource_security_group.InboundRulesValue, len(*http.InboundRules))
	for i, e := range *http.InboundRules {
		ibds[i] = InboundRuleFromHttpToTf(e)
	}

	obds := make([]resource_security_group.OutboundRulesValue, len(*http.OutboundRules))
	for i, e := range *http.OutboundRules {
		obds[i] = OutboundRuleFromHttpToTf(ctx, e)
	}

	ibdsTf, diag := types.ListValueFrom(ctx, resource_security_group.InboundRulesValue{}.Type(ctx), ibds)
	if diag.HasError() {
		return nil, diag
	}

	obdsType := resource_security_group.OutboundRulesValue{}.Type(ctx)
	obdsTf, diag := types.ListValueFrom(ctx, obdsType, obds)
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

func CreateInboundRulesRequest(ctx context.Context, data resource_security_group.SecurityGroupModel, res *api.CreateSecurityGroupResponse) api.CreateSecurityGroupRuleJSONRequestBody {
	tfInboundRules := make([]resource_security_group.InboundRulesValue, 0, len(data.InboundRules.Elements()))
	data.InboundRules.ElementsAs(ctx, &tfInboundRules, false)
	inboundRules := []api.SecurityGroupRuleSchema{}
	for _, e := range tfInboundRules {
		tfIpRange := make([]types.String, 0, len(e.IpRanges.Elements()))
		e.IpRanges.ElementsAs(ctx, &tfIpRange, false)
		var ipRanges []string
		for _, ip := range tfIpRange {
			ipRanges = append(ipRanges, ip.ValueString())
		}

		inboundRules = append(inboundRules, api.SecurityGroupRuleSchema{
			FromPortRange: utils.FromTfInt64ToIntPtr(e.FromPortRange),
			ToPortRange:   utils.FromTfInt64ToIntPtr(e.ToPortRange),
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			IpRanges:      &ipRanges,
		})
	}

	inboundRulesCreationBody := api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:            "Inbound",
		SecurityGroupId: *res.JSON200.Id,
		Rules:           &inboundRules,
	}
	return inboundRulesCreationBody
}

func CreateOutboundRulesRequest(ctx context.Context, data resource_security_group.SecurityGroupModel, res *api.CreateSecurityGroupResponse) api.CreateSecurityGroupRuleJSONRequestBody {
	tfInboundRules := make([]resource_security_group.OutboundRulesValue, 0, len(data.InboundRules.Elements()))
	data.InboundRules.ElementsAs(ctx, &tfInboundRules, false)
	outboundRules := []api.SecurityGroupRuleSchema{}
	for _, e := range tfInboundRules {
		tfIpRange := make([]types.String, 0, len(e.IpRanges.Elements()))
		e.IpRanges.ElementsAs(ctx, &tfIpRange, false)
		var ipRanges []string
		for _, ip := range tfIpRange {
			ipRanges = append(ipRanges, ip.ValueString())
		}

		schema := api.SecurityGroupRuleSchema{
			FromPortRange: utils.FromTfInt64ToIntPtr(e.FromPortRange),
			ToPortRange:   utils.FromTfInt64ToIntPtr(e.ToPortRange),
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			IpRanges:      &ipRanges,
		}

		outboundRules = append(outboundRules, schema)
	}

	outboundRulesCreationBody := api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:            "Inbound",
		SecurityGroupId: *res.JSON200.Id,
		Rules:           &outboundRules,
	}
	return outboundRulesCreationBody
}
