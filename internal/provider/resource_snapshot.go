package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_snapshot"
)

var (
	_ resource.Resource                = &SnapshotResource{}
	_ resource.ResourceWithConfigure   = &SnapshotResource{}
	_ resource.ResourceWithImportState = &SnapshotResource{}
)

type SnapshotResource struct {
	client *api.ClientWithResponses
}

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

func (r *SnapshotResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *SnapshotResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *SnapshotResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_snapshot"
}

func (r *SnapshotResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_snapshot.SnapshotResourceSchema(ctx)
}

func (r *SnapshotResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	body := SnapshotFromTfToCreateRequest(data)
	res, err := r.client.CreateSnapshotWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create Snapshot", err.Error())
	}

	expectedStatusCode := 201 //FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to create Snapshot", "My Custom Error")
		return
	}

	tf := SnapshotFromHttpToTf(res.JSON201) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SnapshotResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement READ operation
	res, err := r.client.ReadSnapshotsByIdWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to read Snapshot", "My Custom Error")
		return
	}

	tf := SnapshotFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SnapshotResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *SnapshotResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeleteSnapshotWithResponse(ctx, data.Id.String())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete Snapshot", err.Error())
		return
	}

	expectedStatusCode := 204 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		response.Diagnostics.AddError("Failed to delete Snapshot", "My Custom Error")
		return
	}
}
