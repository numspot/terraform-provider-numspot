package volume

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VolumeResource{}
	_ resource.ResourceWithConfigure   = &VolumeResource{}
	_ resource.ResourceWithImportState = &VolumeResource{}
)

type VolumeResource struct {
	provider services.IProvider
}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

func (r *VolumeResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *VolumeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VolumeResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VolumeResourceSchema(ctx)
}

func (r *VolumeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data VolumeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		VolumeFromTfToCreateRequest(&data),
		r.provider.GetNumspotClient().CreateVolumeWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Volume", err.Error())
		return
	}

	// Retries read on resource until state is OK
	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Retries read on resource until state is OK
	read, err := utils2.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"creating"},
		[]string{"available"},
		r.provider.GetNumspotClient().ReadVolumesByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create volume", fmt.Sprintf("Error waiting for volume (%s) to be created: %s", *res.JSON201.Id, err))
		return
	}

	rr, ok := read.(*numspot.Volume)
	if !ok {
		response.Diagnostics.AddError("Failed to create volume", "object conversion error")
		return
	}
	tf, diags := VolumeFromHttpToTf(ctx, rr)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VolumeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils2.ExecuteRequest(func() (*numspot.ReadVolumesByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVolumesByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := VolumeFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VolumeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VolumeModel
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
	}

	updatedRes := utils2.ExecuteRequest(func() (*numspot.UpdateVolumeResponse, error) {
		body := ValueFromTfToUpdaterequest(&plan)
		return r.provider.GetNumspotClient().UpdateVolumeWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString(), body)
	}, http.StatusOK, &response.Diagnostics)
	if updatedRes == nil {
		return
	}

	time.Sleep(3 * time.Second) // TODO remove when outscale fixes the State field => https://numsproduct.atlassian.net/browse/CLSEXP-612

	volumeId := state.Id.ValueString()
	// Retries read on resource until state is OK
	read, err := utils2.RetryReadUntilStateValid(
		ctx,
		volumeId,
		r.provider.GetSpaceID(),
		[]string{"creating", "updating"},
		[]string{"available"},
		r.provider.GetNumspotClient().ReadVolumesByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to update volume", fmt.Sprintf("Error waiting for volume (%s) to be created: %s", state.Id.ValueString(), err))
		return
	}

	rr, ok := read.(*numspot.Volume)
	if !ok {
		response.Diagnostics.AddError("Failed to update volume", "object conversion error")
		return
	}

	tf, diags := VolumeFromHttpToTf(ctx, rr)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VolumeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteVolumeWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Volume", err.Error())
		return
	}
}
