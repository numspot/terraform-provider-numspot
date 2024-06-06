package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_image"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
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
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		*ImageFromTfToCreateRequest(ctx, &data, &response.Diagnostics),
		r.provider.ApiClient.CreateImageWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Image", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.ApiClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Retries read on resource until state is OK
	waitedImage, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"available"},
		r.provider.ApiClient.ReadImagesByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Image", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	image, ok := waitedImage.(*iaas.Image)
	if !ok {
		response.Diagnostics.AddError("Failed to create Image", fmt.Sprintf("Error waiting for instance (%s) to be created", createdId))
	}

	tf, diagnostics := ImageFromHttpToTf(ctx, image)
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
	var state, plan resource_image.ImageModel
	modifications := false

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.ApiClient,
			r.provider.SpaceID,
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}
		modifications = true
	}

	if !modifications {
		return
	}

	res := utils.ExecuteRequest(func() (*iaas.ReadImagesByIdResponse, error) {
		return r.provider.ApiClient.ReadImagesByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)

	tf, diagnostics := ImageFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ImageResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_image.ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteImageWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Image", err.Error())
		return
	}
}
