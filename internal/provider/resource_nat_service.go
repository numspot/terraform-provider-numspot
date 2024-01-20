package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nat_service"
)

var (
	_ resource.Resource                = &NatServiceResource{}
	_ resource.ResourceWithConfigure   = &NatServiceResource{}
	_ resource.ResourceWithImportState = &NatServiceResource{}
)

type NatServiceResource struct {
	client *api.ClientWithResponses
}

func NewNatServiceResource() resource.Resource {
	return &NatServiceResource{}
}

func (r *NatServiceResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NatServiceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NatServiceResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nat_service"
}

func (r *NatServiceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_nat_service.NatServiceResourceSchema(ctx)
}

func (r *NatServiceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_nat_service.NatServiceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	body := NatServiceFromTfToCreateRequest(data)
	res, err := r.client.CreateNatServiceWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create NatService", err.Error())
	}

	expectedStatusCode := 201 //FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to create NatService", "My Custom Error")
		return
	}

	tf := NatServiceFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NatServiceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_nat_service.NatServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement READ operation
	res, err := r.client.ReadNatServicesByIdWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to read NatService", "My Custom Error")
		return
	}

	tf := NatServiceFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NatServiceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *NatServiceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_nat_service.NatServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeleteNatServiceWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete NatService", err.Error())
		return
	}

	expectedStatusCode := 204 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete NatService", "My Custom Error")
		return
	}
}
