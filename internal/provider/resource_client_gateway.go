package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_client_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &ClientGatewayResource{}
	_ resource.ResourceWithConfigure   = &ClientGatewayResource{}
	_ resource.ResourceWithImportState = &ClientGatewayResource{}
)

type ClientGatewayResource struct {
	provider Provider
}

func NewClientGatewayResource() resource.Resource {
	return &ClientGatewayResource{}
}

func (r *ClientGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		ClientGatewayFromTfToCreateRequest(&data),
		r.provider.ApiClient.CreateClientGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Client Gateway", err.Error())
		return
	}

	// Retries read on resource until state is OK
	createdId := *res.JSON201.Id
	_, err = retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"available"},
		r.provider.ApiClient.ReadClientGatewaysByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Client Gateways", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	tf := ClientGatewayFromHttpToTf(res.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *ClientGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadClientGatewaysByIdResponse, error) {
		return r.provider.ApiClient.ReadClientGatewaysByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := ClientGatewayFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *ClientGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *ClientGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteClientGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Client Gateway", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
