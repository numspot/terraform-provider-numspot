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

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
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

func (r *SecurityGroupResource) deleteRules(ctx context.Context, id string, existingRules *[]numspot.SecurityGroupRule, flow string, diags *diag.Diagnostics) {
	if existingRules == nil {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
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

	utils.ExecuteRequest(func() (*numspot.DeleteSecurityGroupRuleResponse, error) {
		body := numspot.DeleteSecurityGroupRuleJSONRequestBody{
			Flow:  flow,
			Rules: &rules,
		}
		return numspotClient.DeleteSecurityGroupRuleWithResponse(ctx, r.provider.SpaceID, id, body)
	}, http.StatusNoContent, diags)
}

// Note : this is not a method of SecurityGroupResource because method do not handle generic types
func createRules[RulesType any](
	r *SecurityGroupResource,
	ctx context.Context,
	id string,
	rulesToCreate basetypes.SetValue,
	fun func(ctx context.Context, sgId string, data []RulesType) numspot.CreateSecurityGroupRuleJSONRequestBody,
	diags *diag.Diagnostics,
) {
	rules := make([]RulesType, 0, len(rulesToCreate.Elements()))
	rulesToCreate.ElementsAs(ctx, &rules, false)
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	_ = utils.ExecuteRequest(func() (*numspot.CreateSecurityGroupRuleResponse, error) {
		body := fun(ctx, id, rules)
		return numspotClient.CreateSecurityGroupRuleWithResponse(ctx, r.provider.SpaceID, id, body)
	}, http.StatusCreated, diags)
}

func (r *SecurityGroupResource) updateAllRules(ctx context.Context, data SecurityGroupModel, id string, diags *diag.Diagnostics) {
	// Read security group to retrieve the existing rules
	read := r.readSecurityGroup(ctx, id, diags)
	if diags.HasError() {
		return
	}

	// Delete existing inbound rules
	if len(*read.JSON200.InboundRules) > 0 {
		r.deleteRules(ctx, id, read.JSON200.InboundRules, "Inbound", diags)
		if diags.HasError() {
			return
		}
	}

	// Create wanted inbound rules
	if len(data.InboundRules.Elements()) > 0 {
		createRules(r, ctx, id, data.InboundRules, CreateInboundRulesRequest, diags)
		if diags.HasError() {
			return
		}
	}

	// Delete existing Outbound rules
	if len(*read.JSON200.OutboundRules) > 0 {
		r.deleteRules(ctx, id, read.JSON200.OutboundRules, "Outbound", diags)
		if diags.HasError() {
			return
		}
	}
	// Create wanted Outbound rules
	if len(data.OutboundRules.Elements()) > 0 {
		createRules(r, ctx, id, data.OutboundRules, CreateOutboundRulesRequest, diags)
		if diags.HasError() {
			return
		}
	}
}

func (r *SecurityGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data SecurityGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		SecurityGroupFromTfToCreateRequest(&data),
		numspotClient.CreateSecurityGroupWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Security Group", err.Error())
		return
	}

	id := utils.GetPtrValue(res.JSON201.Id)
	if id == "" {
		return
	}

	// Create tags
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, numspotClient, r.provider.SpaceID, &response.Diagnostics, id, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	r.updateAllRules(ctx, data, id, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	// Read before store
	read := r.readSecurityGroup(ctx, id, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	tf := SecurityGroupFromHttpToTf(ctx, read.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	res := r.readSecurityGroup(ctx, data.Id.ValueString(), &response.Diagnostics)
	if response.Diagnostics.HasError() || res == nil {
		return
	}

	tf := SecurityGroupFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) readSecurityGroup(
	ctx context.Context,
	id string,
	diagnostics *diag.Diagnostics,
) *numspot.ReadSecurityGroupsByIdResponse {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}

	res, err := numspotClient.ReadSecurityGroupsByIdWithResponse(ctx, r.provider.SpaceID, id)
	if err != nil {
		diagnostics.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
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

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
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
			numspotClient,
			r.provider.SpaceID,
			securityGroupId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// update rules
	r.updateAllRules(ctx, plan, securityGroupId, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	res := r.readSecurityGroup(ctx, securityGroupId, &response.Diagnostics)
	if response.Diagnostics.HasError() || res == nil {
		return
	}

	tf := SecurityGroupFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), numspotClient.DeleteSecurityGroupWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Security Group", err.Error())
		return
	}
}
