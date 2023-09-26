package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link_interface"
)

var (
	_ resource.Resource                = &DirectLinkInterfaceResource{}
	_ resource.ResourceWithConfigure   = &DirectLinkInterfaceResource{}
	_ resource.ResourceWithImportState = &DirectLinkInterfaceResource{}
)

type DirectLinkInterfaceResource struct {
	client *api.ClientWithResponses
}

func NewDirectLinkInterfaceResource() resource.Resource {
	return &DirectLinkInterfaceResource{}
}

func (r *DirectLinkInterfaceResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *DirectLinkInterfaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *DirectLinkInterfaceResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_direct_link_interface"
}

func (r *DirectLinkInterfaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_direct_link_interface.DirectLinkInterfaceResourceSchema(ctx)
}

func (r *DirectLinkInterfaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_direct_link_interface.DirectLinkInterfaceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	body := DirectLinkInterfaceFromTfToCreateRequest(data)
	res, err := r.client.CreateDirectLinkInterfaceWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create DirectLinkInterface", err.Error())
	}

	expectedStatusCode := 201 //FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to create DirectLinkInterface", "My Custom Error")
		return
	}

	tf := DirectLinkInterfaceFromHttpToTf(res.JSON201) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkInterfaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_direct_link_interface.DirectLinkInterfaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement READ operation
	res, err := r.client.ReadDirectLinkInterfacesByIdWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to read DirectLinkInterface", "My Custom Error")
		return
	}

	tf := DirectLinkInterfaceFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkInterfaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *DirectLinkInterfaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_direct_link_interface.DirectLinkInterfaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeleteDirectLinkInterfaceWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete DirectLinkInterface", err.Error())
		return
	}

	expectedStatusCode := 204 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete DirectLinkInterface", "My Custom Error")
		return
	}
}
