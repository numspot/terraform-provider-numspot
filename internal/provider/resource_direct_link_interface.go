package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link_interface"
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

	res := utils.ExecuteRequest(func() (*iaas.CreateDirectLinkInterfaceResponse, error) {
		body := DirectLinkInterfaceFromTfToCreateRequest(&data)
		return r.provider.ApiClient.CreateDirectLinkInterfaceWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := DirectLinkInterfaceFromHttpToTf(res.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkInterfaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_direct_link_interface.DirectLinkInterfaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadDirectLinkInterfacesByIdResponse, error) {
		return r.provider.ApiClient.ReadDirectLinkInterfacesByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
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

	utils.ExecuteRequest(func() (*iaas.DeleteDirectLinkInterfaceResponse, error) {
		return r.provider.ApiClient.DeleteDirectLinkInterfaceWithResponse(ctx, r.provider.SpaceID, data.Id.String())
	}, http.StatusOK, &response.Diagnostics)
}
