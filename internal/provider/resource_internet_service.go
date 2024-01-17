package provider

import (
  "context"
  "fmt"

  "github.com/hashicorp/terraform-plugin-framework/path"
  "github.com/hashicorp/terraform-plugin-framework/resource"

  "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
  "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_service"
)

var _ resource.Resource = &InternetServiceResource{}
var _ resource.ResourceWithConfigure = &InternetServiceResource{}
var _ resource.ResourceWithImportState = &InternetServiceResource{}

type InternetServiceResource struct{
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

  body := InternetServiceFromTfToCreateRequest(data)
  res, err := r.client.CreateInternetServiceWithResponse(ctx, body)
  if err != nil {
    // TODO: Handle Error
    response.Diagnostics.AddError("Failed to create InternetService", err.Error())
  }

  expectedStatusCode := 201 //FIXME: Set expected status code (must be 201)
  if res.StatusCode() != expectedStatusCode {
    // TODO: Handle NumSpot error
    response.Diagnostics.AddError("Failed to create InternetService", "My Custom Error")
    return
  }


  tf := InternetServiceFromHttpToTf(res.JSON201) // FIXME
  response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *InternetServiceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
  var data resource_internet_service.InternetServiceModel
  response.Diagnostics.Append(request.State.Get(ctx, &data)...)

  //TODO: Implement READ operation
  res, err := r.client.ReadInternetServicesByIdWithResponse(ctx, data.Id.String())
  if err != nil {
    // TODO: Handle Error
    response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
  }

  expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
  if res.StatusCode() != expectedStatusCode {
    // TODO: Handle NumSpot error
    response.Diagnostics.AddError("Failed to read InternetService", "My Custom Error")
    return
  }

  tf := InternetServiceFromHttpToTf(res.JSON200) // FIXME
  response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *InternetServiceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
  //TODO implement me
  panic("implement me")
}

func (r *InternetServiceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_internet_service.InternetServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement DELETE operation
	res, err := r.client.DeleteInternetServiceWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete InternetService", err.Error())
		return
	}

	expectedStatusCode := 204 //FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete InternetService", "My Custom Error")
		return
	}
}
