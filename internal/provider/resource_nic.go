package provider

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nic"
)

var (
	_ resource.Resource                = &NicResource{}
	_ resource.ResourceWithConfigure   = &NicResource{}
	_ resource.ResourceWithImportState = &NicResource{}
)

type NicResource struct {
	provider Provider
}

func NewNicResource() resource.Resource {
	return &NicResource{}
}

func (r *NicResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NicResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NicResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nic"
}

func (r *NicResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_nic.NicResourceSchema(ctx)
}

func (r *NicResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_nic.NicModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.CreateNicResponse, error) {
		body := NicFromTfToCreateRequest(ctx, &data)
		return r.provider.ApiClient.CreateNicWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := NicFromHttpToTf(ctx, res.JSON201)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NicResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_nic.NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadNicsByIdResponse, error) {
		return r.provider.ApiClient.ReadNicsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := NicFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NicResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *NicResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_nic.NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	_ = utils.ExecuteRequest(func() (*api.DeleteNicResponse, error) {
		return r.provider.ApiClient.DeleteNicWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
}
