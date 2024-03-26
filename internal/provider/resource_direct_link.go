package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &DirectLinkResource{}
	_ resource.ResourceWithConfigure   = &DirectLinkResource{}
	_ resource.ResourceWithImportState = &DirectLinkResource{}
)

type DirectLinkResource struct {
	provider Provider
}

func NewDirectLinkResource() resource.Resource {
	return &DirectLinkResource{}
}

func (r *DirectLinkResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *DirectLinkResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *DirectLinkResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_direct_link"
}

func (r *DirectLinkResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_direct_link.DirectLinkResourceSchema(ctx)
}

func (r *DirectLinkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_direct_link.DirectLinkModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.CreateDirectLinkResponse, error) {
		body := DirectLinkFromTfToCreateRequest(&data)
		return r.provider.ApiClient.CreateDirectLinkWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := DirectLinkFromHttpToTf(res.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_direct_link.DirectLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadDirectLinksByIdResponse, error) {
		return r.provider.ApiClient.ReadDirectLinksByIdWithResponse(ctx, r.provider.SpaceID, data.Id.String())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := DirectLinkFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *DirectLinkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_direct_link.DirectLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.DeleteDirectLinkResponse, error) {
		return r.provider.ApiClient.DeleteDirectLinkWithResponse(ctx, r.provider.SpaceID, data.Id.String())
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}
}
