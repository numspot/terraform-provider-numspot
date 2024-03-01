package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

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

	res := utils.ExecuteRequest(func() (*api.CreateSnapshotResponse, error) {
		body := SnapshotFromTfToCreateRequest(&data)
		return r.client.CreateSnapshotWithResponse(ctx, spaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := SnapshotFromHttpToTf(res.JSON201)
	if !data.SourceRegionName.IsUnknown() {
		tf.SourceRegionName = data.SourceRegionName
	} else {
		tf.SourceRegionName = types.StringNull()
	}

	if !data.SourceSnapshotId.IsUnknown() {
		tf.SourceSnapshotId = data.SourceSnapshotId
	} else {
		tf.SourceSnapshotId = types.StringNull()
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SnapshotResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadSnapshotsByIdResponse, error) {
		return r.client.ReadSnapshotsByIdWithResponse(ctx, spaceID, data.Id.String())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := SnapshotFromHttpToTf(res.JSON200)
	if !data.SourceRegionName.IsUnknown() {
		tf.SourceRegionName = data.SourceRegionName
	} else {
		tf.SourceRegionName = types.StringNull()
	}

	if !data.SourceSnapshotId.IsUnknown() {
		tf.SourceSnapshotId = data.SourceSnapshotId
	} else {
		tf.SourceSnapshotId = types.StringNull()
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SnapshotResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *SnapshotResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	_ = utils.ExecuteRequest(func() (*api.DeleteSnapshotResponse, error) {
		return r.client.DeleteSnapshotWithResponse(ctx, spaceID, data.Id.String())
	}, http.StatusNoContent, &response.Diagnostics)
}
