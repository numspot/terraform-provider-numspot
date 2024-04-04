package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &InternetGatewayResource{}
	_ resource.ResourceWithConfigure   = &InternetGatewayResource{}
	_ resource.ResourceWithImportState = &InternetGatewayResource{}
)

type InternetGatewayResource struct {
	provider Provider
}

func NewInternetGatewayResource() resource.Resource {
	return &InternetGatewayResource{}
}

func (r *InternetGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *InternetGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *InternetGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_internet_gateway"
}

func (r *InternetGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_internet_gateway.InternetGatewayResourceSchema(ctx)
}

func (r *InternetGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.CreateInternetGatewayResponse, error) {
		return r.provider.ApiClient.CreateInternetGatewayWithResponse(ctx, r.provider.SpaceID)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil || res.JSON201 == nil {
		return
	}

	createdId := res.JSON201.Id

	// Call Link Internet Service to VPC
	vpcId := data.VpcIp
	if !vpcId.IsNull() {
		linRes := utils.ExecuteRequest(func() (*iaas.LinkInternetGatewayResponse, error) {
			return r.provider.ApiClient.LinkInternetGatewayWithResponse(
				ctx,
				r.provider.SpaceID,
				*createdId,
				iaas.LinkInternetGatewayJSONRequestBody{
					VpcId: data.VpcIp.ValueString(),
				},
			)
		}, http.StatusNoContent, &response.Diagnostics)
		if linRes == nil {
			return
		}
	}

	// Update state
	readRes := utils.ExecuteRequest(func() (*iaas.ReadInternetGatewaysByIdResponse, error) {
		return r.provider.ApiClient.ReadInternetGatewaysByIdWithResponse(ctx, r.provider.SpaceID, *createdId)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := InternetServiceFromHttpToTf(readRes.JSON200)
	tf.VpcIp = data.VpcIp
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *InternetGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadInternetGatewaysByIdResponse, error) {
		return r.provider.ApiClient.ReadInternetGatewaysByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := InternetServiceFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *InternetGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *InternetGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if !data.VpcIp.IsNull() {
		res := utils.ExecuteRequest(func() (*iaas.UnlinkInternetGatewayResponse, error) {
			return r.provider.ApiClient.UnlinkInternetGatewayWithResponse(
				ctx,
				r.provider.SpaceID,
				data.Id.ValueString(),
				iaas.UnlinkInternetGatewayJSONRequestBody{
					VpcId: data.VpcIp.ValueString(),
				})
		}, http.StatusNoContent, &response.Diagnostics)
		if res == nil {
			return
		}
	}

	res, err := r.provider.ApiClient.DeleteInternetGatewayWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Failed to delete InternetService", err.Error())
		return
	}

	if res.StatusCode() != http.StatusNoContent {
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to delete InternetService", apiError.Error())
		return
	}
}
