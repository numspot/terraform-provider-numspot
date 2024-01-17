package provider

import (
  "context"
  "fmt"

  "github.com/hashicorp/terraform-plugin-framework/path"
  "github.com/hashicorp/terraform-plugin-framework/resource"

  "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
  "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_net_peering"
)

var _ resource.Resource = &NetPeeringResource{}
var _ resource.ResourceWithConfigure = &NetPeeringResource{}
var _ resource.ResourceWithImportState = &NetPeeringResource{}

type NetPeeringResource struct{
  client *api.ClientWithResponses
}

func NewNetPeeringResource() resource.Resource {
	return &NetPeeringResource{}
}

func (r *NetPeeringResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NetPeeringResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
  resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NetPeeringResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
  response.TypeName = request.ProviderTypeName + "_net_peering"
}

func (r *NetPeeringResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
  response.Schema = resource_net_peering.NetPeeringResourceSchema(ctx)
}

func (r *NetPeeringResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
  var data resource_net_peering.NetPeeringModel
  response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

  body := NetPeeringFromTfToCreateRequest(data)
  res, err := r.client.CreateNetPeeringWithResponse(ctx, body)
  if err != nil {
    // TODO: Handle Error
    response.Diagnostics.AddError("Failed to create NetPeering", err.Error())
  }

  expectedStatusCode := 201 //FIXME: Set expected status code (must be 201)
  if res.StatusCode() != expectedStatusCode {
    // TODO: Handle NumSpot error
    response.Diagnostics.AddError("Failed to create NetPeering", "My Custom Error")
    return
  }


  tf := NetPeeringFromHttpToTf(res.JSON201) // FIXME
  response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NetPeeringResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
  var data resource_net_peering.NetPeeringModel
  response.Diagnostics.Append(request.State.Get(ctx, &data)...)

  //TODO: Implement READ operation
  res, err := r.client.ReadNetPeeringsByIdWithResponse(ctx, data.Id.String())
  if err != nil {
    // TODO: Handle Error
    response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
  }

  expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
  if res.StatusCode() != expectedStatusCode {
    // TODO: Handle NumSpot error
    response.Diagnostics.AddError("Failed to read NetPeering", "My Custom Error")
    return
  }

  tf := NetPeeringFromHttpToTf(res.JSON200) // FIXME
  response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NetPeeringResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
  //TODO implement me
  panic("implement me")
}

func (r *NetPeeringResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_net_peering.NetPeeringModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement DELETE operation
	res, err := r.client.DeleteNetPeeringWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete NetPeering", err.Error())
		return
	}

	expectedStatusCode := 204 //FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete NetPeering", "My Custom Error")
		return
	}
}
