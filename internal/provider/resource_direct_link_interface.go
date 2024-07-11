package provider

/*

 DIRECT LINKS are not handled for now

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link_interface"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &DirectLinkInterfaceResource{}
	_ resource.ResourceWithConfigure   = &DirectLinkInterfaceResource{}
	_ resource.ResourceWithImportState = &DirectLinkInterfaceResource{}
)

type DirectLinkInterfaceResource struct {
	provider Provider
}

func NewDirectLinkInterfaceResource() resource.Resource {
	return &DirectLinkInterfaceResource{}
}

func (r *DirectLinkInterfaceResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *DirectLinkInterfaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *DirectLinkInterfaceResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_direct_link_interface"
}

func (r *DirectLinkInterfaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_direct_link_interface.DirectLinkInterfaceResourceSchema(ctx)
}

func (r *DirectLinkInterfaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_direct_link_interface.DirectLinkInterfaceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		DirectLinkInterfaceFromTfToCreateRequest(&data),
		r.provider.IaasClient.CreateDirectLinkInterfaceWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Direct Link Interface", err.Error())
		return
	}

	// Retries read on resource until state is OK
	createdId := *res.JSON201.DirectLinkId
	_, err = retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending", "confirming"},
		[]string{"available"},
		r.provider.IaasClient.ReadDirectLinkInterfacesByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Direct Link Interface", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	tf := DirectLinkInterfaceFromHttpToTf(res.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkInterfaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_direct_link_interface.DirectLinkInterfaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*numspot.ReadDirectLinkInterfacesByIdResponse, error) {
		return r.provider.IaasClient.ReadDirectLinkInterfacesByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := DirectLinkInterfaceFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkInterfaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *DirectLinkInterfaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_direct_link_interface.DirectLinkInterfaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.IaasClient.DeleteDirectLinkInterfaceWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Direct Link Interface", err.Error())
		return
	}

	utils.ExecuteRequest(func() (*numspot.DeleteDirectLinkInterfaceResponse, error) {
		return r.provider.IaasClient.DeleteDirectLinkInterfaceWithResponse(ctx, r.provider.SpaceID, data.Id.String())
	}, http.StatusOK, &response.Diagnostics)
}
*/
