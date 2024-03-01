package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"net/http"
	"time"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_volume"
)

var (
	_ resource.Resource                = &VolumeResource{}
	_ resource.ResourceWithConfigure   = &VolumeResource{}
	_ resource.ResourceWithImportState = &VolumeResource{}
)

type VolumeResource struct {
	client *api.ClientWithResponses
}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

func (r *VolumeResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *VolumeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VolumeResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_volume.VolumeResourceSchema(ctx)
}

func (r *VolumeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_volume.VolumeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.CreateVolumeResponse, error) {
		body := VolumeFromTfToCreateRequest(&data)
		return r.client.CreateVolumeWithResponse(ctx, spaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	createStateConf := &retry.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"available"},
		Refresh: func() (result interface{}, state string, err error) {
			readRes := utils.ExecuteRequest(func() (*api.ReadVolumesByIdResponse, error) {
				return r.client.ReadVolumesByIdWithResponse(ctx, *res.JSON201.Id)
			}, http.StatusOK, &response.Diagnostics)
			if readRes == nil {
				return
			}

			return readRes.JSON200, *readRes.JSON200.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   3 * time.Second,
	}

	read, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for volume (%s) to be created: %s", *res.JSON201.Id, err))
		return
	}

	rr := read.(*api.Volume)
	tf, diags := VolumeFromHttpToTf(ctx, rr)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VolumeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_volume.VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadVolumesByIdResponse, error) {
		return r.client.ReadVolumesByIdWithResponse(ctx, spaceID, data.Id.String())
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
	var state, plan resource_volume.VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	updatedRes := utils.ExecuteRequest(func() (*api.UpdateVolumeResponse, error) {
		body := ValueFromTfToUpdaterequest(&plan)
		return r.client.UpdateVolumeWithResponse(ctx, spaceID, state.Id.ValueString(), body)
	}, http.StatusOK, &response.Diagnostics)
	if updatedRes == nil {
		return
	}

	volumeId := state.Id.ValueString()
	updateStateConf := &retry.StateChangeConf{
		Pending: []string{"creating", "updating"},
		Target:  []string{"available"},
		Refresh: func() (result interface{}, state string, err error) {
			readRes := utils.ExecuteRequest(func() (*api.ReadVolumesByIdResponse, error) {
				return r.client.ReadVolumesByIdWithResponse(ctx, volumeId)
			}, http.StatusOK, &response.Diagnostics)
			if readRes == nil {
				return
			}

			return readRes.JSON200, *readRes.JSON200.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   3 * time.Second,
	}

	read, err := updateStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for volume (%s) to be created: %s", state.Id.ValueString(), err))
		return
	}

	rr := read.(*api.Volume)
	tf, diags := VolumeFromHttpToTf(ctx, rr)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VolumeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_volume.VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	utils.ExecuteRequest(func() (*api.DeleteVolumeResponse, error) {
		return r.client.DeleteVolumeWithResponse(ctx, spaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
}
