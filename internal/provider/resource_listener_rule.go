package provider

import (
  "context"
  "fmt"

  "github.com/hashicorp/terraform-plugin-framework/path"
  "github.com/hashicorp/terraform-plugin-framework/resource"

  "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
  "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_listener_rule"
)

var _ resource.Resource = &ListenerRuleResource{}
var _ resource.ResourceWithConfigure = &ListenerRuleResource{}
var _ resource.ResourceWithImportState = &ListenerRuleResource{}

type ListenerRuleResource struct{
  client *api.ClientWithResponses
}

func NewListenerRuleResource() resource.Resource {
	return &ListenerRuleResource{}
}

func (r *ListenerRuleResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *ListenerRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
  resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *ListenerRuleResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
  response.TypeName = request.ProviderTypeName + "_listener_rule"
}

func (r *ListenerRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
  response.Schema = resource_listener_rule.ListenerRuleResourceSchema(ctx)
}

func (r *ListenerRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
  var data resource_listener_rule.ListenerRuleModel
  response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

  body := ListenerRuleFromTfToCreateRequest(data)
  res, err := r.client.CreateListenerRuleWithResponse(ctx, body)
  if err != nil {
    // TODO: Handle Error
    response.Diagnostics.AddError("Failed to create ListenerRule", err.Error())
  }

  expectedStatusCode := 201 //FIXME: Set expected status code (must be 201)
  if res.StatusCode() != expectedStatusCode {
    // TODO: Handle NumSpot error
    response.Diagnostics.AddError("Failed to create ListenerRule", "My Custom Error")
    return
  }


  tf := ListenerRuleFromHttpToTf(res.JSON201) // FIXME
  response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ListenerRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
  var data resource_listener_rule.ListenerRuleModel
  response.Diagnostics.Append(request.State.Get(ctx, &data)...)

  //TODO: Implement READ operation
  res, err := r.client.ReadListenerRulesByIdWithResponse(ctx, data.Id.String())
  if err != nil {
    // TODO: Handle Error
    response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
  }

  expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
  if res.StatusCode() != expectedStatusCode {
    // TODO: Handle NumSpot error
    response.Diagnostics.AddError("Failed to read ListenerRule", "My Custom Error")
    return
  }

  tf := ListenerRuleFromHttpToTf(res.JSON200) // FIXME
  response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ListenerRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
  //TODO implement me
  panic("implement me")
}

func (r *ListenerRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_listener_rule.ListenerRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement DELETE operation
	res, err := r.client.DeleteListenerRuleWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete ListenerRule", err.Error())
		return
	}

	expectedStatusCode := 204 //FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete ListenerRule", "My Custom Error")
		return
	}
}
