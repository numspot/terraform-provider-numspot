package provider

import (
	"context"
	"fmt"

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

	body := SecurityGroupFromTfToCreateRequest(data)
	res, err := r.client.CreateSecurityGroupWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create SecurityGroup", err.Error())
	}

	expectedStatusCode := 201 // FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to create SecurityGroup", "My Custom Error")
		return
	}

	// Inbound
	inboundRulesCreationBody := CreateInboundRulesRequest(ctx, data, res)
	createInboundRulesResponse, err := r.client.CreateSecurityGroupRuleWithResponse(ctx, inboundRulesCreationBody)
	if err != nil {
		response.Diagnostics.AddError("Failed to create SecurityGroup", err.Error())
		return
	}
	fmt.Println(createInboundRulesResponse)

	// Outbound
	outboundRulesCreationBody := CreateOutboundRulesRequest(ctx, data, res)
	createOutboundRulesResponse, err := r.client.CreateSecurityGroupRuleWithResponse(ctx, outboundRulesCreationBody)
	if err != nil {
		response.Diagnostics.AddError("Failed to create SecurityGroup", err.Error())
		return
	}
	fmt.Println(createOutboundRulesResponse)

	// Read before store
	tf := SecurityGroupFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement READ operation
	res, err := r.client.ReadSecurityGroupsByIdWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 // FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to read SecurityGroup", "My Custom Error")
		return
	}

	tf := SecurityGroupFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SecurityGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {

	panic("implement me")
}

func (r *SecurityGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeleteSecurityGroupWithResponse(ctx, data.Id.String(), api.DeleteSecurityGroupRequestSchema{})
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete SecurityGroup", err.Error())
		return
	}

	expectedStatusCode := 204 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete SecurityGroup", "My Custom Error")
		return
	}
}
