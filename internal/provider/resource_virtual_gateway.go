package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_virtual_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VirtualGatewayResource{}
	_ resource.ResourceWithConfigure   = &VirtualGatewayResource{}
	_ resource.ResourceWithImportState = &VirtualGatewayResource{}
)

type VirtualGatewayResource struct {
	provider Provider
}

func NewVirtualGatewayResource() resource.Resource {
	return &VirtualGatewayResource{}
}

func (r *VirtualGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = client
}

func (r *VirtualGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VirtualGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_virtual_gateway"
}

func (r *VirtualGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_virtual_gateway.VirtualGatewayResourceSchema(ctx)
}

func (r *VirtualGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		VirtualGatewayFromTfToCreateRequest(data),
		r.provider.ApiClient.CreateVirtualGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Virtual Gateway", err.Error())
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
		r.provider.ApiClient.ReadVirtualGatewaysByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Virtual Gateway", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	tf, diagnostics := VirtualGatewayFromHttpToTf(ctx, res.JSON201)
	if diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VirtualGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadVirtualGatewaysByIdResponse, error) {
		return r.provider.ApiClient.ReadVirtualGatewaysByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := VirtualGatewayFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VirtualGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *VirtualGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteVirtualGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Virtual Gateway", err.Error())
		return
	}
}
