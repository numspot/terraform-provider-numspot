package provider

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_image"
)

var (
	_ resource.Resource                = &ImageResource{}
	_ resource.ResourceWithConfigure   = &ImageResource{}
	_ resource.ResourceWithImportState = &ImageResource{}
)

type ImageResource struct {
	client *api.ClientWithResponses
}

func NewImageResource() resource.Resource {
	return &ImageResource{}
}

func (r *ImageResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *ImageResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *ImageResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_image"
}

func (r *ImageResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_image.ImageResourceSchema(ctx)
}

func (r *ImageResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_image.ImageModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.CreateImageResponse, error) {
		body := ImageFromTfToCreateRequest(&data)
		return r.client.CreateImageWithResponse(ctx, body)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := ImageFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ImageResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_image.ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadImagesByIdResponse, error) {
		return r.client.ReadImagesByIdWithResponse(ctx, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)

	tf := ImageFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ImageResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *ImageResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_image.ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	utils.ExecuteRequest(func() (*api.DeleteImageResponse, error) {
		return r.client.DeleteImageWithResponse(ctx, data.Id.ValueString(), api.DeleteImageJSONRequestBody{})
	}, http.StatusOK, &response.Diagnostics)
}
