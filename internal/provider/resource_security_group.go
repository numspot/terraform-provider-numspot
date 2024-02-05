package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_security_group"
)

var (
	_ resource.Resource                = &SecurityGroupResource{}
	_ resource.ResourceWithConfigure   = &SecurityGroupResource{}
	_ resource.ResourceWithImportState = &SecurityGroupResource{}
)

type SecurityGroupResource struct {
	client *api.ClientWithResponses
}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*api.ClientWithResponses)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.client = client
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

	res := utils.HandleResponse(func() (*api.CreateSecurityGroupResponse, error) {
		body := SecurityGroupFromTfToCreateRequest(&data)
		return r.client.CreateSecurityGroupWithResponse(ctx, body)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	createdId := res.JSON200.Id

	// Inbound
	if len(data.InboundRules.Elements()) > 0 {
		inboundRules := make([]resource_security_group.InboundRulesValue, 0, len(data.InboundRules.Elements()))
		data.InboundRules.ElementsAs(ctx, &inboundRules, false)

		createdIbdRules := utils.HandleResponse(func() (*api.CreateSecurityGroupRuleResponse, error) {
			body := CreateInboundRulesRequest(ctx, *createdId, inboundRules)
			return r.client.CreateSecurityGroupRuleWithResponse(ctx, body)
		}, http.StatusOK, &response.Diagnostics)
		if createdIbdRules == nil {
			return
		}
	}

	// Outbound
	if len(data.OutboundRules.Elements()) > 0 {
		// Delete default SG rule: (0.0.0.0/0 -1)
		hRules := *res.JSON200.OutboundRules
		rules := make([]api.SecurityGroupRuleSchema, 0, len(hRules))
		for _, e := range hRules {
			rules = append(rules, api.SecurityGroupRuleSchema{
				FromPortRange:         e.FromPortRange,
				IpProtocol:            e.IpProtocol,
				IpRanges:              e.IpRanges,
				SecurityGroupsMembers: e.SecurityGroupsMembers,
				ServiceIds:            e.ServiceIds,
				ToPortRange:           e.ToPortRange,
			})
		}

		utils.HandleResponse(func() (*api.DeleteSecurityGroupRuleResponse, error) {
			body := api.DeleteSecurityGroupRuleJSONRequestBody{
				Flow:  "Outbound",
				Rules: &rules,
			}
			return r.client.DeleteSecurityGroupRuleWithResponse(ctx, *createdId, body)
		}, http.StatusOK, &response.Diagnostics)

		// Create SG rules:
		outboundRules := make([]resource_security_group.OutboundRulesValue, 0, len(data.OutboundRules.Elements()))
		data.OutboundRules.ElementsAs(ctx, &outboundRules, false)

		createdObdRules := utils.HandleResponse(func() (*api.CreateSecurityGroupRuleResponse, error) {
			body := CreateOutboundRulesRequest(ctx, *createdId, outboundRules)
			return r.client.CreateSecurityGroupRuleWithResponse(ctx, body)
		}, http.StatusOK, &response.Diagnostics)
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
	if response.Diagnostics.HasError() {
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
) *api.ReadSecurityGroupsByIdResponse {
	res, err := r.client.ReadSecurityGroupsByIdWithResponse(ctx, id)
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

	_ = utils.HandleResponse(func() (*api.DeleteSecurityGroupResponse, error) {
		return r.client.DeleteSecurityGroupWithResponse(ctx, data.Id.ValueString(), api.DeleteSecurityGroupRequestSchema{})
	}, http.StatusOK, &response.Diagnostics)
}
