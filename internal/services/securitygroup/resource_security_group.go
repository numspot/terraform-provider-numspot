package securitygroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SecurityGroupResource{}
	_ resource.ResourceWithConfigure   = &SecurityGroupResource{}
	_ resource.ResourceWithImportState = &SecurityGroupResource{}
)

type SecurityGroupResource struct {
	provider *client.NumSpotSDK
}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *SecurityGroupResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *SecurityGroupResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_security_group"
}

func (r *SecurityGroupResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = SecurityGroupResourceSchema(ctx)
}

func (r *SecurityGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan SecurityGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsList := tags.TfTagsToApiTags(ctx, plan.Tags)
	inboundRules := deserializeCreateInboundRules(ctx, plan.InboundRules)
	outboundRules := deserializeCreateOutboundRules(ctx, plan.OutboundRules)

	numSpotSecurityGroup, err := core.CreateSecurityGroup(ctx, r.provider, deserializeCreateSecurityGroupRequest(plan), tagsList, &inboundRules, &outboundRules)
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

func (r *SecurityGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state SecurityGroupModel
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

func (r *SecurityGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan          SecurityGroupModel
		err                  error
		numSpotSecurityGroup *numspot.SecurityGroup
	)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	if !plan.Tags.Equal(state.Tags) {
		numSpotSecurityGroup, err = core.UpdateSecurityGroupTags(ctx, r.provider, state.Id.ValueString(), stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update security group tags", err.Error())
			return
		}
	}

	planInboundRules := deserializeCreateInboundRules(ctx, plan.InboundRules)
	planOutboundRules := deserializeCreateOutboundRules(ctx, plan.OutboundRules)
	stateInboundRules := deserializeDeleteInboundRules(ctx, state.InboundRules)
	stateOutboundRules := deserializeDeleteOutboundRules(ctx, state.OutboundRules)

	if !plan.InboundRules.Equal(state.InboundRules) || !plan.OutboundRules.Equal(state.OutboundRules) {
		numSpotSecurityGroup, err = core.UpdateSecurityGroupRules(ctx, r.provider, state.Id.ValueString(), planInboundRules, planOutboundRules, &stateInboundRules, &stateOutboundRules)
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

func (r *SecurityGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state SecurityGroupModel
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

func serializeSecurityGroup(ctx context.Context, http *numspot.SecurityGroup, diags *diag.Diagnostics) *SecurityGroupModel {
	var tagsTf types.List

	if http.InboundRules == nil {
		return nil
	}
	ibd := make([]InboundRulesValue, 0, len(*http.InboundRules))
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
	obd := make([]OutboundRulesValue, 0, len(*http.OutboundRules))
	for _, e := range *http.OutboundRules {
		value := serializeOutboundRule(ctx, e)
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

func deserializeCreateSecurityGroupRequest(tf SecurityGroupModel) numspot.CreateSecurityGroupJSONRequestBody {
	return numspot.CreateSecurityGroupJSONRequestBody{
		Description: tf.Description.ValueString(),
		VpcId:       tf.VpcId.ValueString(),
		Name:        tf.Name.ValueString(),
	}
}

func deserializeCreateInboundRules(ctx context.Context, inboundRules types.Set) numspot.CreateSecurityGroupRuleJSONRequestBody {
	rules := deserializeInboundRules(ctx, inboundRules)

	inboundRulesCreationBody := numspot.CreateSecurityGroupRuleJSONRequestBody{
		Flow:  "Inbound",
		Rules: &rules,
	}

	return inboundRulesCreationBody
}

func deserializeCreateOutboundRules(ctx context.Context, outboundRules types.Set) numspot.CreateSecurityGroupRuleJSONRequestBody {
	rules := deserializeOutboundRules(ctx, outboundRules)

	outboundRulesCreationBody := numspot.CreateSecurityGroupRuleJSONRequestBody{
		Flow:  "Outbound",
		Rules: &rules,
	}

	return outboundRulesCreationBody
}

func deserializeDeleteInboundRules(ctx context.Context, inboundRules types.Set) numspot.DeleteSecurityGroupRuleJSONRequestBody {
	rules := deserializeInboundRules(ctx, inboundRules)

	return numspot.DeleteSecurityGroupRuleJSONRequestBody{
		Flow:  "Inbound",
		Rules: &rules,
	}
}

func deserializeDeleteOutboundRules(ctx context.Context, outboundRules types.Set) numspot.DeleteSecurityGroupRuleJSONRequestBody {
	rules := deserializeOutboundRules(ctx, outboundRules)

	return numspot.DeleteSecurityGroupRuleJSONRequestBody{
		Flow:  "Outbound",
		Rules: &rules,
	}
}

func deserializeInboundRules(ctx context.Context, inboundRules types.Set) []numspot.SecurityGroupRule {
	tfRules := make([]InboundRulesValue, 0, len(inboundRules.Elements()))
	inboundRules.ElementsAs(ctx, &tfRules, false)
	rules := make([]numspot.SecurityGroupRule, 0, len(tfRules))
	for i := range tfRules {
		e := &tfRules[i]
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

	return rules
}

func deserializeOutboundRules(ctx context.Context, outboundRules types.Set) []numspot.SecurityGroupRule {
	tfRules := make([]OutboundRulesValue, 0, len(outboundRules.Elements()))
	outboundRules.ElementsAs(ctx, &tfRules, false)
	rules := make([]numspot.SecurityGroupRule, 0, len(tfRules))
	for i := range tfRules {
		e := &tfRules[i]

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

	return rules
}

func serializeInboundRule(ctx context.Context, rules numspot.SecurityGroupRule) InboundRulesValue {
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

func serializeOutboundRule(ctx context.Context, rules numspot.SecurityGroupRule) OutboundRulesValue {
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
