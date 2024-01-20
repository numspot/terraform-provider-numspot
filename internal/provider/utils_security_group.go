package provider

import (
	"context"

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

func SecurityGroupFromHttpToTf(http *api.SecurityGroupSchema) resource_security_group.SecurityGroupModel {
	return resource_security_group.SecurityGroupModel{
		AccountId:     types.StringPointerValue(http.AccountId),
		Description:   types.StringPointerValue(http.Description),
		Id:            types.StringPointerValue(http.Id),
		InboundRules:  types.List{},
		Name:          types.StringPointerValue(http.Name),
		NetId:         types.StringPointerValue(http.NetId),
		OutboundRules: types.List{},
	}
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
