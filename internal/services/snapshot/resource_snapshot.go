package snapshot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SnapshotResource{}
	_ resource.ResourceWithConfigure   = &SnapshotResource{}
	_ resource.ResourceWithImportState = &SnapshotResource{}
)

type SnapshotResource struct {
	provider services.IProvider
}

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

func (r *SnapshotResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
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
	response.Schema = SnapshotResourceSchema(ctx)
}

func (r *SnapshotResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data SnapshotModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		SnapshotFromTfToCreateRequest(&data),
		r.provider.GetNumspotClient().CreateSnapshotWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Snapshot", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Retries read on resource until state is OK
	readRes, err := utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"pending/queued", "in-queue", "pending"},
		[]string{"completed"},
		r.provider.GetNumspotClient().ReadSnapshotsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Snapshot", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	rr, ok := readRes.(*numspot.Snapshot)
	if !ok {
		response.Diagnostics.AddError("Failed to create Snapshot", "object conversion error")
		return
	}

	tf := SnapshotFromHttpToTf(ctx, rr, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

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
	var data SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*numspot.ReadSnapshotsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadSnapshotsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := SnapshotFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

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
	var state, plan SnapshotModel
	modifications := false

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}

		modifications = true
	}

	if !modifications {
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadSnapshotsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadSnapshotsByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := SnapshotFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.SourceRegionName.IsUnknown() {
		tf.SourceRegionName = state.SourceRegionName
	} else {
		tf.SourceRegionName = types.StringNull()
	}

	if !state.SourceSnapshotId.IsUnknown() {
		tf.SourceSnapshotId = state.SourceSnapshotId
	} else {
		tf.SourceSnapshotId = types.StringNull()
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SnapshotResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data SnapshotModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteSnapshotWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Snapshot", err.Error())
		return
	}
}
