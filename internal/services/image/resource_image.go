package image

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &ImageResource{}
	_ resource.ResourceWithConfigure   = &ImageResource{}
	_ resource.ResourceWithImportState = &ImageResource{}
)

type ImageResource struct {
	provider services.IProvider
}

func NewImageResource() resource.Resource {
	return &ImageResource{}
}

func (r *ImageResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
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
	response.Schema = ImageResourceSchema(ctx)
}

func (r *ImageResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ImageModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		*ImageFromTfToCreateRequest(ctx, &data, &response.Diagnostics),
		r.provider.GetNumspotClient().CreateImageWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Image", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if !utils.IsTfValueNull(data.Access) {
		response.Diagnostics.Append(r.updateImageAccess(ctx, createdId, data)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Retries read on resource until state is OK
	waitedImage, err := utils2.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"pending"},
		[]string{"available"},
		r.provider.GetNumspotClient().ReadImagesByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Image", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	image, ok := waitedImage.(*numspot.Image)
	if !ok {
		response.Diagnostics.AddError("Failed to create Image", fmt.Sprintf("Error waiting for instance (%s) to be created", createdId))
	}

	if response.Diagnostics.HasError() {
		return
	}

	tf, diags := r.parseImageObjectToTf(ctx, data, *image)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ImageResource) parseImageObjectToTf(ctx context.Context, data ImageModel, image numspot.Image) (*ImageModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	tf, diagnostics := ImageFromHttpToTf(ctx, &image)
	if diagnostics.HasError() {
		diags.Append(diagnostics...)
		return nil, diags
	}

	tf.SourceImageId = utils2.FromTfStringValueToTfOrNull(data.SourceImageId)
	tf.SourceRegionName = utils2.FromTfStringValueToTfOrNull(data.SourceRegionName)
	tf.VmId = utils2.FromTfStringValueToTfOrNull(data.VmId)
	tf.NoReboot = utils2.FromTfBoolValueToTfOrNull(data.NoReboot)

	return tf, nil
}

func (r *ImageResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ImageModel
	var diags diag.Diagnostics

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils2.ExecuteRequest(func() (*numspot.ReadImagesByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadImagesByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	tf, diags := r.parseImageObjectToTf(ctx, data, *res.JSON200)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ImageResource) updateImageAccess(ctx context.Context, id string, data ImageModel) diag.Diagnostics {
	var body numspot.UpdateImageJSONRequestBody
	var diags diag.Diagnostics
	if data.Access.IsPublic.ValueBool() { // If IsPublic is set to True
		body = numspot.UpdateImageJSONRequestBody{
			AccessCreation: numspot.AccessCreation{
				Additions: &numspot.Access{
					IsPublic: utils.EmptyTrueBoolPointer(),
				},
				Removals: nil,
			},
		}
	} else { // If IsPublic is set to False or removed
		body = numspot.UpdateImageJSONRequestBody{
			AccessCreation: numspot.AccessCreation{
				Additions: nil,
				Removals: &numspot.Access{
					IsPublic: utils.EmptyTrueBoolPointer(),
				},
			},
		}
	}

	_ = utils2.ExecuteRequest(func() (*numspot.UpdateImageResponse, error) {
		return r.provider.GetNumspotClient().UpdateImageWithResponse(ctx,
			r.provider.GetSpaceID(),
			id,
			body,
		)
	}, http.StatusOK, &diags)

	return diags
}

func (r *ImageResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan ImageModel
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
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}
		modifications = true
	}

	if !state.Access.Equal(plan.Access) {
		response.Diagnostics.Append(r.updateImageAccess(ctx, state.Id.ValueString(), plan)...)
		if response.Diagnostics.HasError() {
			return
		}
		modifications = true
	}

	if !modifications {
		return
	}

	res := utils2.ExecuteRequest(func() (*numspot.ReadImagesByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadImagesByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	tf, diags := r.parseImageObjectToTf(ctx, state, *res.JSON200)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ImageResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ImageModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteImageWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Image", err.Error())
		return
	}
}
