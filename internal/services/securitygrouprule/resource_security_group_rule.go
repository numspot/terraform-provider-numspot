package securitygrouprule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/securitygrouprule/resource_security_group_rule"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &securityGroupRuleResource{}
	_ resource.ResourceWithConfigure   = &securityGroupRuleResource{}
	_ resource.ResourceWithImportState = &securityGroupRuleResource{}
)

// securityGroupRuleResource is the Terraform resource implementation for numspot_security_group_rule.
type securityGroupRuleResource struct {
	services.BaseResource
}

// NewSecurityGroupRuleResource returns a new securityGroupRuleResource instance.
func NewSecurityGroupRuleResource() resource.Resource {
	return &securityGroupRuleResource{}
}

// Metadata sets the resource type name.
func (r *securityGroupRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_security_group_rule"
}

// Schema defines the schema for the resource.
func (r *securityGroupRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_security_group_rule.SecurityGroupRuleResourceSchema(ctx)
}

// ImportState imports the resource state.
func (r *securityGroupRuleResource) ImportState(_ context.Context, _ resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.AddError("Import not supported", "the numspot_security_group_rule resource does not support import")
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *securityGroupRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_security_group_rule.SecurityGroupRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotSecurityGroupRule, err := core.CreateSecurityGroupRule(
		ctx,
		r.Provider,
		deserializeCreateSecurityGroupRuleRequest(&plan),
		plan.SecurityGroupId.ValueString(),
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create security_group_rule", err.Error())
		return
	}

	state := serializeSecurityGroupRule(numspotSecurityGroupRule)
	state.Flow = utils.PreserveFromState(plan.Flow)
	state.SecurityGroupId = utils.PreserveFromState(plan.SecurityGroupId)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *securityGroupRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_security_group_rule.SecurityGroupRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotSecurityGroupRule, err := core.ReadSecurityGroupRule(ctx, r.Provider, state.SecurityGroupId.ValueString(), state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read security_group_rule", err.Error())
		return
	}

	newState := serializeSecurityGroupRule(numspotSecurityGroupRule)
	newState.Flow = state.Flow
	newState.SecurityGroupId = state.SecurityGroupId

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *securityGroupRuleResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *securityGroupRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_security_group_rule.SecurityGroupRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteSecurityGroupRule(
		ctx,
		r.Provider,
		state.SecurityGroupId.ValueString(),
		state.Id.ValueString(),
		state.Flow.ValueString(),
	); err != nil {
		response.Diagnostics.AddError("unable to delete security_group_rule", err.Error())
		return
	}
}

func serializeSecurityGroupRule(http *api.SecurityGroupRule) resource_security_group_rule.SecurityGroupRuleModel {
	return resource_security_group_rule.SecurityGroupRuleModel{
		FromPortRange: utils.FromIntPtrToTfInt64(http.FromPortRange),
		Id:            securityGroupRuleID(http),
		IpProtocol:    types.StringPointerValue(http.IpProtocol),
		IpRange:       firstIpRange(http.IpRanges),
		ToPortRange:   utils.FromIntPtrToTfInt64(http.ToPortRange),
	}
}

func deserializeCreateSecurityGroupRuleRequest(tf *resource_security_group_rule.SecurityGroupRuleModel) *api.CreateSecurityGroupRuleJSONRequestBody {
	return &api.CreateSecurityGroupRuleJSONRequestBody{
		Flow:          tf.Flow.ValueString(),
		FromPortRange: utils.FromTfInt64ToIntPtr(tf.FromPortRange),
		IpProtocol:    tf.IpProtocol.ValueStringPointer(),
		IpRange:       tf.IpRange.ValueStringPointer(),
		ToPortRange:   utils.FromTfInt64ToIntPtr(tf.ToPortRange),
	}
}
