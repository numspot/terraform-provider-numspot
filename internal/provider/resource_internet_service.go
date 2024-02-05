package provider

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_service"
)

var (
	_ resource.Resource                = &InternetServiceResource{}
	_ resource.ResourceWithConfigure   = &InternetServiceResource{}
	_ resource.ResourceWithImportState = &InternetServiceResource{}
)

type InternetServiceResource struct {
	client *api.ClientWithResponses
}

func NewInternetServiceResource() resource.Resource {
	return &InternetServiceResource{}
}

func (r *InternetServiceResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *InternetServiceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *InternetServiceResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_internet_service"
}

func (r *InternetServiceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_internet_service.InternetServiceResourceSchema(ctx)
}

func (r *InternetServiceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_internet_service.InternetServiceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res, err := r.client.CreateInternetServiceWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create InternetService", err.Error())
		return
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to create InternetService", apiError.Error())
		return
	}
	// Update state
	tf := InternetServiceFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)

	// Call Link Internet Service to VPC
	netID := data.NetId
	if !netID.IsNull() {
		reslink, errlink := r.client.LinkInternetServiceWithResponse(
			ctx,
			tf.Id.ValueString(),
			api.LinkInternetServiceJSONRequestBody{
				NetId: data.NetId.ValueString(),
			})
		if errlink != nil {
			response.Diagnostics.AddError("Failed to link InternetService to net", err.Error())
			return
		}
		expectedStatusCode := 200
		if reslink.StatusCode() != expectedStatusCode {
			apiError := utils.HandleError(res.Body)
			response.Diagnostics.AddError("Failed to link InternetService to net", apiError.Error())
			return
		}
		// Update state
		tf.NetId = data.NetId
		response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
	}
}

func (r *InternetServiceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_internet_service.InternetServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res, err := r.client.ReadInternetServicesByIdWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Failed to read Internet service", err.Error())
		return
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to read InternetService", apiError.Error())
		return
	}

	tf := InternetServiceFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *InternetServiceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *InternetServiceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_internet_service.InternetServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if !data.NetId.IsNull() {
		res, err := r.client.UnlinkInternetServiceWithResponse(
			ctx,
			data.Id.ValueString(),
			api.UnlinkInternetServiceJSONRequestBody{
				NetId: data.NetId.ValueString(),
			})
		if err != nil {
			response.Diagnostics.AddError("Failed to unlink InternetService from net", err.Error())
			return
		}

		if res.StatusCode() != http.StatusOK {
			apiError := utils.HandleError(res.Body)
			response.Diagnostics.AddError("Failed to unlink InternetService from net", apiError.Error())
			return
		}
	}

	res, err := r.client.DeleteInternetServiceWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Failed to delete InternetService", err.Error())
		return
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to delete InternetService", apiError.Error())
		return
	}
}
