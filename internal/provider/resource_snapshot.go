package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_snapshot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SnapshotResource{}
	_ resource.ResourceWithConfigure   = &SnapshotResource{}
	_ resource.ResourceWithImportState = &SnapshotResource{}
)

type SnapshotResource struct {
	provider Provider
}

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

func (r *SnapshotResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
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

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		SnapshotFromTfToCreateRequest(&data),
		r.provider.ApiClient.CreateSnapshotWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Snapshot", err.Error())
		return
	}

	// Retries read on resource until state is OK
	createdId := *res.JSON201.Id
	_, err = retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending/queued", "in-queue", "pending"},
		[]string{"completed"},
		r.provider.ApiClient.ReadSnapshotsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Snapshot", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
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

	res := utils.ExecuteRequest(func() (*iaas.ReadSnapshotsByIdResponse, error) {
		return r.provider.ApiClient.ReadSnapshotsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
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
	panic("nothing to do")
}

func (r *SnapshotResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_snapshot.SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteSnapshotWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Snapshot", err.Error())
		return
	}
}
