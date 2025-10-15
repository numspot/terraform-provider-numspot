package securitygroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/securitygroup/resource_security_group"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &securityGroupResource{}
	_ resource.ResourceWithConfigure   = &securityGroupResource{}
	_ resource.ResourceWithImportState = &securityGroupResource{}
)

type securityGroupResource struct {
	provider *client.NumSpotSDK
}

func NewSecurityGroupResource() resource.Resource {
	return &securityGroupResource{}
}

func (r *securityGroupResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *securityGroupResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *securityGroupResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_security_group"
}

func (r *securityGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_security_group.SecurityGroupResourceSchema(ctx)
}

func (r *securityGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_security_group.SecurityGroupModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsList := securityGroupTags(ctx, plan.Tags)
	inboundRules := deserializeCreateInboundRules(ctx, plan.InboundRules)
	outboundRules := deserializeCreateOutboundRules(ctx, plan.OutboundRules)

	numSpotSecurityGroup, err := core.CreateSecurityGroup(ctx, r.provider, deserializeCreateSecurityGroupRequest(plan), tagsList, inboundRules, outboundRules)
	if err != nil {
		response.Diagnostics.AddError("unable to create security group", err.Error())
		return
	}

	state := serializeSecurityGroup(ctx, numSpotSecurityGroup, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_security_group.SecurityGroupModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	securityGroupID := state.Id.ValueString()

	numSpotSecurityGroup, err := core.ReadSecurityGroup(ctx, r.provider, securityGroupID)
	if err != nil {
		response.Diagnostics.AddError("unable to read security group", err.Error())
		return
	}

	newState := serializeSecurityGroup(ctx, numSpotSecurityGroup, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *securityGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan          resource_security_group.SecurityGroupModel
		err                  error
		numSpotSecurityGroup *api.SecurityGroup
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	stateTags := securityGroupTags(ctx, state.Tags)
	planTags := securityGroupTags(ctx, plan.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotSecurityGroup, err = core.UpdateSecurityGroupTags(ctx, r.provider, state.Id.ValueString(), stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update security group tags", err.Error())
			return
		}
	}

	securityGroupID := state.Id.ValueString()
	planInboundRules := deserializeCreateInboundRules(ctx, plan.InboundRules)
	planOutboundRules := deserializeCreateOutboundRules(ctx, plan.OutboundRules)
	stateInboundRules := deserializeDeleteInboundRules(ctx, state.InboundRules)
	stateOutboundRules := deserializeDeleteOutboundRules(ctx, state.OutboundRules)

	if !plan.InboundRules.Equal(state.InboundRules) || !plan.OutboundRules.Equal(state.OutboundRules) {
		numSpotSecurityGroup, err = core.UpdateSecurityGroupRules(ctx, r.provider, securityGroupID, stateInboundRules, stateOutboundRules, planInboundRules, planOutboundRules)
		if err != nil {
			response.Diagnostics.AddError("unable to update security group rules", err.Error())
		}
	}

	newState := serializeSecurityGroup(ctx, numSpotSecurityGroup, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *securityGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	securityGroupID := state.Id.ValueString()

	if err := core.DeleteSecurityGroup(ctx, r.provider, securityGroupID); err != nil {
		response.Diagnostics.AddError("unable to delete security group", err.Error())
		return
	}
}

func serializeSecurityGroup(ctx context.Context, http *api.SecurityGroup, diags *diag.Diagnostics) *resource_security_group.SecurityGroupModel {
	var tagsTf types.Set

	if http.InboundRules == nil {
		return nil
	}
	ibd := make([]resource_security_group.InboundRulesValue, 0, len(*http.InboundRules))
	for _, e := range *http.InboundRules {
		value := serializeInboundRule(ctx, e)
		if diags.HasError() {
			return nil
		}

		ibd = append(ibd, value)
	}

	if http.OutboundRules == nil {
		return nil
	}
	obd := make([]resource_security_group.OutboundRulesValue, 0, len(*http.OutboundRules))
	for _, e := range *http.OutboundRules {
		value := serializeOutboundRule(ctx, e)
		if diags.HasError() {
			return nil
		}

		obd = append(obd, value)
	}

	ibdsTf, diagnostics := types.SetValueFrom(ctx, resource_security_group.InboundRulesValue{}.Type(ctx), ibd)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return nil
	}

	obdsTf, diagnostics := types.SetValueFrom(ctx, resource_security_group.OutboundRulesValue{}.Type(ctx), obd)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return nil
	}

	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	res := resource_security_group.SecurityGroupModel{
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

func deserializeCreateSecurityGroupRequest(tf resource_security_group.SecurityGroupModel) api.CreateSecurityGroupJSONRequestBody {
	return api.CreateSecurityGroupJSONRequestBody{
		Description: tf.Description.ValueString(),
		VpcId:       tf.VpcId.ValueString(),
		Name:        tf.Name.ValueString(),
	}
}

func deserializeCreateInboundRules(ctx context.Context, inboundRules types.Set) api.CreateSecurityGroupRuleJSONRequestBody {
	rules := deserializeInboundRules(ctx, inboundRules)

	inboundRulesCreationBody := api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:  "Inbound",
		Rules: &rules,
	}

	return inboundRulesCreationBody
}

func deserializeCreateOutboundRules(ctx context.Context, outboundRules types.Set) api.CreateSecurityGroupRuleJSONRequestBody {
	rules := deserializeOutboundRules(ctx, outboundRules)

	outboundRulesCreationBody := api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:  "Outbound",
		Rules: &rules,
	}

	return outboundRulesCreationBody
}

func deserializeDeleteInboundRules(ctx context.Context, inboundRules types.Set) api.DeleteSecurityGroupRuleJSONRequestBody {
	rules := deserializeInboundRules(ctx, inboundRules)

	return api.DeleteSecurityGroupRuleJSONRequestBody{
		Flow:  "Inbound",
		Rules: &rules,
	}
}

func deserializeDeleteOutboundRules(ctx context.Context, outboundRules types.Set) api.DeleteSecurityGroupRuleJSONRequestBody {
	rules := deserializeOutboundRules(ctx, outboundRules)

	return api.DeleteSecurityGroupRuleJSONRequestBody{
		Flow:  "Outbound",
		Rules: &rules,
	}
}

func deserializeInboundRules(ctx context.Context, inboundRules types.Set) []api.SecurityGroupRule {
	tfRules := make([]resource_security_group.InboundRulesValue, 0, len(inboundRules.Elements()))
	inboundRules.ElementsAs(ctx, &tfRules, false)
	rules := make([]api.SecurityGroupRule, 0, len(tfRules))
	for i := range tfRules {
		e := &tfRules[i]
		fpr := int(e.FromPortRange.ValueInt64())
		tpr := int(e.ToPortRange.ValueInt64())

		rules = append(rules, api.SecurityGroupRule{
			FromPortRange: &fpr,
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			ToPortRange:   &tpr,
		})
	}

	return rules
}

func deserializeOutboundRules(ctx context.Context, outboundRules types.Set) []api.SecurityGroupRule {
	tfRules := make([]resource_security_group.OutboundRulesValue, 0, len(outboundRules.Elements()))
	outboundRules.ElementsAs(ctx, &tfRules, false)
	rules := make([]api.SecurityGroupRule, 0, len(tfRules))
	for i := range tfRules {
		e := &tfRules[i]

		fpr := int(e.FromPortRange.ValueInt64())
		tpr := int(e.ToPortRange.ValueInt64())

		rules = append(rules, api.SecurityGroupRule{
			FromPortRange: &fpr,
			IpProtocol:    e.IpProtocol.ValueStringPointer(),
			ToPortRange:   &tpr,
		})
	}

	return rules
}

func serializeInboundRule(ctx context.Context, rules api.SecurityGroupRule) resource_security_group.InboundRulesValue {
	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return resource_security_group.InboundRulesValue{}
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return resource_security_group.InboundRulesValue{}
		}
	}

	value, diagnostics := resource_security_group.NewInboundRulesValue(
		resource_security_group.InboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range": utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":   utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":     types.StringPointerValue(rules.IpProtocol),
			"service_ids":     serviceIds,
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeOutboundRule(ctx context.Context, rules api.SecurityGroupRule) resource_security_group.OutboundRulesValue {
	serviceIds, diags := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	if diags.HasError() {
		return resource_security_group.OutboundRulesValue{}
	}

	if rules.ServiceIds == nil {
		serviceIds, diags = types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			return resource_security_group.OutboundRulesValue{}
		}
	}

	value, diagnostics := resource_security_group.NewOutboundRulesValue(
		resource_security_group.OutboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range": utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":   utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":     types.StringPointerValue(rules.IpProtocol),
			"service_ids":     serviceIds,
		},
	)
	diags.Append(diagnostics...)
	return value
}

func securityGroupTags(ctx context.Context, tags types.Set) []api.ResourceTag {
	tfTags := make([]resource_security_group.TagsValue, 0, len(tags.Elements()))
	tags.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
