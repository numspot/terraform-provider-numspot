package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_security_group"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SecurityGroupResource{}
	_ resource.ResourceWithConfigure   = &SecurityGroupResource{}
	_ resource.ResourceWithImportState = &SecurityGroupResource{}
)

type SecurityGroupResource struct {
	provider Provider
}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
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
	response.Schema = resource_security_group.SecurityGroupResourceSchema(ctx)
}

func (r *SecurityGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		SecurityGroupFromTfToCreateRequest(&data),
		r.provider.ApiClient.CreateSecurityGroupWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Security Group", err.Error())
		return
	}

	createdId := res.JSON201.Id

	// Inbound
	if len(data.InboundRules.Elements()) > 0 {
		inboundRules := make([]resource_security_group.InboundRulesValue, 0, len(data.InboundRules.Elements()))
		data.InboundRules.ElementsAs(ctx, &inboundRules, false)

		createdIbdRules := utils.ExecuteRequest(func() (*iaas.CreateSecurityGroupRuleResponse, error) {
			body := CreateInboundRulesRequest(ctx, *createdId, inboundRules)
			return r.provider.ApiClient.CreateSecurityGroupRuleWithResponse(ctx, r.provider.SpaceID, *createdId, body)
		}, http.StatusCreated, &response.Diagnostics)
		if createdIbdRules == nil {
			return
		}
	}

	// Outbound
	if len(data.OutboundRules.Elements()) > 0 {
		// Delete default SG rule: (0.0.0.0/0 -1)
		hRules := *res.JSON201.OutboundRules
		rules := make([]iaas.SecurityGroupRule, 0, len(hRules))
		for _, e := range hRules {
			rules = append(rules, iaas.SecurityGroupRule{
				FromPortRange:         e.FromPortRange,
				IpProtocol:            e.IpProtocol,
				IpRanges:              e.IpRanges,
				SecurityGroupsMembers: e.SecurityGroupsMembers,
				ServiceIds:            e.ServiceIds,
				ToPortRange:           e.ToPortRange,
			})
		}

		utils.ExecuteRequest(func() (*iaas.DeleteSecurityGroupRuleResponse, error) {
			body := iaas.DeleteSecurityGroupRuleJSONRequestBody{
				Flow:  "Outbound",
				Rules: &rules,
			}
			return r.provider.ApiClient.DeleteSecurityGroupRuleWithResponse(ctx, r.provider.SpaceID, *createdId, body)
		}, http.StatusNoContent, &response.Diagnostics)

		// Create SG rules:
		outboundRules := make([]resource_security_group.OutboundRulesValue, 0, len(data.OutboundRules.Elements()))
		data.OutboundRules.ElementsAs(ctx, &outboundRules, false)

		createdObdRules := utils.ExecuteRequest(func() (*iaas.CreateSecurityGroupRuleResponse, error) {
			body := CreateOutboundRulesRequest(ctx, *createdId, outboundRules)
			return r.provider.ApiClient.CreateSecurityGroupRuleWithResponse(ctx, r.provider.SpaceID, *createdId, body)
		}, http.StatusCreated, &response.Diagnostics)
		if createdObdRules == nil {
			return
		}
	}

	// Read before store
	read := r.readSecurityGroup(ctx, *createdId, response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	tf, diagnostics := SecurityGroupFromHttpToTf(ctx, data, read.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := r.readSecurityGroup(ctx, data.Id.ValueString(), response.Diagnostics)
	if response.Diagnostics.HasError() || res == nil {
		return
	}

	tf, diagnostics := SecurityGroupFromHttpToTf(ctx, data, res.JSON200)
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
) *iaas.ReadSecurityGroupsByIdResponse {
	res, err := r.provider.ApiClient.ReadSecurityGroupsByIdWithResponse(ctx, r.provider.SpaceID, id)
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
	panic("implement me")
}

func (r *SecurityGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteSecurityGroupWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Security Group", err.Error())
		return
	}
}
