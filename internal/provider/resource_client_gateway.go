package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_client_gateway"
)

var (
	_ resource.Resource                = &ClientGatewayResource{}
	_ resource.ResourceWithConfigure   = &ClientGatewayResource{}
	_ resource.ResourceWithImportState = &ClientGatewayResource{}
)

type ClientGatewayResource struct {
	client *api.ClientWithResponses
}

func NewClientGatewayResource() resource.Resource {
	return &ClientGatewayResource{}
}

func (r *ClientGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *ClientGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *ClientGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_client_gateway"
}

func (r *ClientGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_client_gateway.ClientGatewayResourceSchema(ctx)
}

func (r *ClientGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	body := ClientGatewayFromTfToCreateRequest(data)
	res, err := r.client.CreateClientGatewayWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create ClientGateway", err.Error())
	}

	expectedStatusCode := 201 //FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to create ClientGateway", "My Custom Error")
		return
	}

	tf := ClientGatewayFromHttpToTf(res.JSON201) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ClientGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement READ operation
	res, err := r.client.ReadClientGatewaysByIdWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to read ClientGateway", "My Custom Error")
		return
	}

	tf := ClientGatewayFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ClientGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *ClientGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeleteClientGatewayWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete ClientGateway", err.Error())
		return
	}

	expectedStatusCode := 204 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete ClientGateway", "My Custom Error")
		return
	}
}
