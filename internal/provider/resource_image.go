package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_image"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &ImageResource{}
	_ resource.ResourceWithConfigure   = &ImageResource{}
	_ resource.ResourceWithImportState = &ImageResource{}
)

type ImageResource struct {
	provider Provider
}

func NewImageResource() resource.Resource {
	return &ImageResource{}
}

func (r *ImageResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	body := ImageFromTfToCreateRequest(ctx, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() || body == nil {
		return
	}

	res := utils.ExecuteRequest(func() (*iaas.CreateImageResponse, error) {
		return r.provider.ApiClient.CreateImageWithResponse(ctx, r.provider.SpaceID, *body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := ImageFromHttpToTf(ctx, res.JSON201)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	// Must set those values to keep it:
	tf.SourceImageId = utils.FromTfStringValueToTfOrNull(data.SourceImageId)
	tf.SourceRegionName = utils.FromTfStringValueToTfOrNull(data.SourceRegionName)
	tf.VmId = utils.FromTfStringValueToTfOrNull(data.VmId)
	tf.NoReboot = utils.FromTfBoolValueToTfOrNull(data.NoReboot)
	//

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ImageResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_image.ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadImagesByIdResponse, error) {
		return r.provider.ApiClient.ReadImagesByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)

	tf, diagnostics := ImageFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ImageResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *ImageResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_image.ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	utils.ExecuteRequest(func() (*iaas.DeleteImageResponse, error) {
		return r.provider.ApiClient.DeleteImageWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
}
