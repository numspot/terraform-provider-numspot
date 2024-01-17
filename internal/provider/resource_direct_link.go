package provider

import (
  "context"
  "fmt"

  "github.com/hashicorp/terraform-plugin-framework/path"
  "github.com/hashicorp/terraform-plugin-framework/resource"

  "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
  "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link"
)

var _ resource.Resource = &DirectLinkResource{}
var _ resource.ResourceWithConfigure = &DirectLinkResource{}
var _ resource.ResourceWithImportState = &DirectLinkResource{}

type DirectLinkResource struct{
  client *api.ClientWithResponses
}

func NewDirectLinkResource() resource.Resource {
	return &DirectLinkResource{}
}

func (r *DirectLinkResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

  body := DirectLinkFromTfToCreateRequest(data)
  res, err := r.client.CreateDirectLinkWithResponse(ctx, body)
  if err != nil {
    // TODO: Handle Error
    response.Diagnostics.AddError("Failed to create DirectLink", err.Error())
  }

  expectedStatusCode := 201 //FIXME: Set expected status code (must be 201)
  if res.StatusCode() != expectedStatusCode {
    // TODO: Handle NumSpot error
    response.Diagnostics.AddError("Failed to create DirectLink", "My Custom Error")
    return
  }


  tf := DirectLinkFromHttpToTf(res.JSON201) // FIXME
  response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
  var data resource_direct_link.DirectLinkModel
  response.Diagnostics.Append(request.State.Get(ctx, &data)...)

  //TODO: Implement READ operation
  res, err := r.client.ReadDirectLinksByIdWithResponse(ctx, data.Id.String())
  if err != nil {
    // TODO: Handle Error
    response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
  }

  expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
  if res.StatusCode() != expectedStatusCode {
    // TODO: Handle NumSpot error
    response.Diagnostics.AddError("Failed to read DirectLink", "My Custom Error")
    return
  }

  tf := DirectLinkFromHttpToTf(res.JSON200) // FIXME
  response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
  //TODO implement me
  panic("implement me")
}

func (r *DirectLinkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_direct_link.DirectLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement DELETE operation
	res, err := r.client.DeleteDirectLinkWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete DirectLink", err.Error())
		return
	}

	expectedStatusCode := 204 //FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete DirectLink", "My Custom Error")
		return
	}
}
