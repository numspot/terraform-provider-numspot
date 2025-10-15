package securitygrouprule

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/securitygrouprule/resource_security_group_rule"
	"terraform-provider-numspot/internal/utils"
)

var _ resource.Resource = &securityGroupRuleResource{}

func NewSecurityGroupRuleResource() resource.Resource {
	return &securityGroupRuleResource{}
}

type securityGroupRuleResource struct {
	provider *client.NumSpotSDK
}

func (r *securityGroupRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *securityGroupRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_rule"
}

func (r *securityGroupRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_security_group_rule.SecurityGroupRuleResourceSchema(ctx)
}

func (r *securityGroupRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_security_group_rule.SecurityGroupRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgId := plan.SecurityGroupId.ValueString()
	body := deserializeCreateSecurityGroupRule(plan)

	res, err := core.CreateSecurityGroupRule(ctx, r.provider, sgId, body)
	if err != nil {
		resp.Diagnostics.AddError("unable to create security group rule", err.Error())
		return
	}

	var createdRule *api.SecurityGroupRule

	switch plan.Flow.ValueString() {
	case "Inbound":
		if res.InboundRules != nil {
			createdRule = findMatchingRule(res.InboundRules, plan)
		}
	case "Outbound":
		if res.OutboundRules != nil {
			createdRule = findMatchingRule(res.OutboundRules, plan)
		}
	default:
		resp.Diagnostics.AddError("invalid flow", fmt.Sprintf("unexpected flow type: %s", plan.Flow.ValueString()))
		return
	}

	if createdRule == nil {
		resp.Diagnostics.AddError("rule not found", "The created rule could not be identified in the API response")
		return
	}

	plan.FromPortRange = utils.FromIntPtrToTfInt64(createdRule.FromPortRange)
	plan.ToPortRange = utils.FromIntPtrToTfInt64(createdRule.ToPortRange)
	plan.IpProtocol = types.StringPointerValue(createdRule.IpProtocol)

	if createdRule.IpRanges != nil && len(*createdRule.IpRanges) > 0 {
		plan.IpRange = types.StringValue((*createdRule.IpRanges)[0])
	} else {
		plan.IpRange = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func deserializeCreateSecurityGroupRule(tf resource_security_group_rule.SecurityGroupRuleModel) api.CreateSecurityGroupRuleJSONRequestBody {
	return api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:          tf.Flow.ValueString(),
		FromPortRange: utils.FromTfInt64ToIntPtr(tf.FromPortRange),
		ToPortRange:   utils.FromTfInt64ToIntPtr(tf.ToPortRange),
		IpProtocol:    tf.IpProtocol.ValueStringPointer(),
		IpRange:       tf.IpRange.ValueStringPointer(),
	}
}

func findMatchingRule(rules *[]api.SecurityGroupRule, plan resource_security_group_rule.SecurityGroupRuleModel) *api.SecurityGroupRule {
	if rules == nil {
		return nil
	}

	planProto := plan.IpProtocol.ValueString()
	planFrom := utils.FromTfInt64ToInt(plan.FromPortRange)
	planTo := utils.FromTfInt64ToInt(plan.ToPortRange)
	planRange := plan.IpRange.ValueString()

	for _, rule := range *rules {
		if rule.IpProtocol == nil || *rule.IpProtocol != planProto {
			continue
		}

		if rule.FromPortRange == nil || rule.ToPortRange == nil {
			continue
		}
		if *rule.FromPortRange != planFrom || *rule.ToPortRange != planTo {
			continue
		}

		if rule.IpRanges == nil || len(*rule.IpRanges) == 0 {
			continue
		}

		for _, ip := range *rule.IpRanges {
			if ip == planRange {
				return &rule
			}
		}
	}

	return nil
}

func (r *securityGroupRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan resource_security_group_rule.SecurityGroupRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgId := plan.SecurityGroupId.ValueString()

	res, err := core.ReadSecurityGroup(ctx, r.provider, sgId)
	if err != nil {
		resp.Diagnostics.AddError("unable to read security group", err.Error())
		return
	}

	var matchedRule *api.SecurityGroupRule

	switch plan.Flow.ValueString() {
	case "Inbound":
		matchedRule = findMatchingRule(res.InboundRules, plan)
	case "Outbound":
		matchedRule = findMatchingRule(res.OutboundRules, plan)
	default:
		resp.Diagnostics.AddError("invalid flow", fmt.Sprintf("unexpected flow type: %s", plan.Flow.ValueString()))
		return
	}

	if matchedRule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	plan.FromPortRange = utils.FromIntPtrToTfInt64(matchedRule.FromPortRange)
	plan.ToPortRange = utils.FromIntPtrToTfInt64(matchedRule.ToPortRange)
	plan.IpProtocol = types.StringPointerValue(matchedRule.IpProtocol)

	if matchedRule.IpRanges != nil && len(*matchedRule.IpRanges) > 0 {
		plan.IpRange = types.StringValue((*matchedRule.IpRanges)[0])
	} else {
		plan.IpRange = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupRuleResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *securityGroupRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_security_group_rule.SecurityGroupRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgId := state.SecurityGroupId.ValueString()

	body := api.DeleteSecurityGroupRuleJSONRequestBody{
		Flow:          state.Flow.ValueString(),
		FromPortRange: utils.FromTfInt64ToIntPtr(state.FromPortRange),
		ToPortRange:   utils.FromTfInt64ToIntPtr(state.ToPortRange),
		IpProtocol:    state.IpProtocol.ValueStringPointer(),
		IpRange:       state.IpRange.ValueStringPointer(),
	}

	if err := core.DeleteSecurityGroupRule(ctx, r.provider, sgId, body); err != nil {
		resp.Diagnostics.AddError("unable to delete security group rule", err.Error())
		return
	}
}
