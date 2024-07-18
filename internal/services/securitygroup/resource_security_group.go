package securitygroup

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SecurityGroupResource{}
	_ resource.ResourceWithConfigure   = &SecurityGroupResource{}
	_ resource.ResourceWithImportState = &SecurityGroupResource{}
)

type SecurityGroupResource struct {
	provider services.IProvider
}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
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

func (r *SecurityGroupResource) deleteRules(ctx context.Context, id string, existingRules *[]numspot.SecurityGroupRule, flow string) diag.Diagnostics {
	var diags diag.Diagnostics

	if existingRules == nil {
		return nil
	}

	rules := make([]numspot.SecurityGroupRule, 0, len(*existingRules))
	for _, e := range *existingRules {
		rules = append(rules, numspot.SecurityGroupRule{
			FromPortRange:         e.FromPortRange,
			IpProtocol:            e.IpProtocol,
			IpRanges:              e.IpRanges,
			SecurityGroupsMembers: e.SecurityGroupsMembers,
			ServiceIds:            e.ServiceIds,
			ToPortRange:           e.ToPortRange,
		})
	}

	utils2.ExecuteRequest(func() (*numspot.DeleteSecurityGroupRuleResponse, error) {
		body := numspot.DeleteSecurityGroupRuleJSONRequestBody{
			Flow:  flow,
			Rules: &rules,
		}
		return r.provider.GetNumspotClient().DeleteSecurityGroupRuleWithResponse(ctx, r.provider.GetSpaceID(), id, body)
	}, http.StatusNoContent, &diags)

	return diags
}

// Note : this is not a method of SecurityGroupResource because method do not handle generic types
func createRules[RulesType any](
	r *SecurityGroupResource,
	ctx context.Context,
	id string,
	rulesToCreate basetypes.SetValue,
	fun func(ctx context.Context, sgId string, data []RulesType) numspot.CreateSecurityGroupRuleJSONRequestBody,
) diag.Diagnostics {
	var diags diag.Diagnostics

	rules := make([]RulesType, 0, len(rulesToCreate.Elements()))
	rulesToCreate.ElementsAs(ctx, &rules, false)

	_ = utils2.ExecuteRequest(func() (*numspot.CreateSecurityGroupRuleResponse, error) {
		body := fun(ctx, id, rules)
		return r.provider.GetNumspotClient().CreateSecurityGroupRuleWithResponse(ctx, r.provider.GetSpaceID(), id, body)
	}, http.StatusCreated, &diags)

	return diags
}

func (r *SecurityGroupResource) updateAllRules(ctx context.Context, data SecurityGroupModel, id string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Read security group to retrieve the existing rules
	read := r.readSecurityGroup(ctx, id, diags)
	if diags.HasError() {
		return diags
	}

	// Delete existing inbound rules
	if len(*read.JSON200.InboundRules) > 0 {
		diags = r.deleteRules(ctx, id, read.JSON200.InboundRules, "Inbound")
		if diags.HasError() {
			return diags
		}
	}

	// Create wanted inbound rules
	if len(data.InboundRules.Elements()) > 0 {
		diags = createRules(r, ctx, id, data.InboundRules, CreateInboundRulesRequest)
		if diags.HasError() {
			return diags
		}
	}

	// Delete existing Outbound rules
	if len(*read.JSON200.OutboundRules) > 0 {
		diags = r.deleteRules(ctx, id, read.JSON200.OutboundRules, "Outbound")
		if diags.HasError() {
			return diags
		}
	}
	// Create wanted Outbound rules
	if len(data.OutboundRules.Elements()) > 0 {
		diags = createRules(r, ctx, id, data.OutboundRules, CreateOutboundRulesRequest)
		if diags.HasError() {
			return diags
		}
	}

	return nil
}

func (r *SecurityGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data SecurityGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		SecurityGroupFromTfToCreateRequest(&data),
		r.provider.GetNumspotClient().CreateSecurityGroupWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Security Group", err.Error())
		return
	}

	id := utils2.GetPtrValue(res.JSON201.Id)
	if id == "" {
		return
	}

	// Create tags
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, id, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Create rules and delete the default one
	diags := r.updateAllRules(ctx, data, id)

	if diags.HasError() {
		return
	}

	// Read before store
	read := r.readSecurityGroup(ctx, id, response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	tf, diagnostics := SecurityGroupFromHttpToTf(ctx, read.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := r.readSecurityGroup(ctx, data.Id.ValueString(), response.Diagnostics)
	if response.Diagnostics.HasError() || res == nil {
		return
	}

	tf, diagnostics := SecurityGroupFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) readSecurityGroup(
	ctx context.Context,
	id string,
	diagnostics diag.Diagnostics,
) *numspot.ReadSecurityGroupsByIdResponse {
	res, err := r.provider.GetNumspotClient().ReadSecurityGroupsByIdWithResponse(ctx, r.provider.GetSpaceID(), id)
	if err != nil {
		diagnostics.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils2.HandleError(res.Body)
		diagnostics.AddError("Failed to read SecurityGroup", apiError.Error())
		return nil
	}

	return res
}

func (r *SecurityGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	securityGroupId := state.Id.ValueString()

	// update tags
	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			securityGroupId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// update rules
	diags := r.updateAllRules(ctx, plan, securityGroupId)
	if diags.HasError() {
		return
	}

	res := r.readSecurityGroup(ctx, securityGroupId, response.Diagnostics)
	if response.Diagnostics.HasError() || res == nil {
		return
	}

	tf, diagnostics := SecurityGroupFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteSecurityGroupWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Security Group", err.Error())
		return
	}
}
