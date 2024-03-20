package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_volume"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VolumeResource{}
	_ resource.ResourceWithConfigure   = &VolumeResource{}
	_ resource.ResourceWithImportState = &VolumeResource{}
)

type VolumeResource struct {
	provider Provider
}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

func (r *VolumeResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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
		return r.provider.ApiClient.CreateVolumeWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	createStateConf := &retry.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"available"},
		Refresh: func() (result interface{}, state string, err error) {
			readRes := utils.ExecuteRequest(func() (*api.ReadVolumesByIdResponse, error) {
				return r.provider.ApiClient.ReadVolumesByIdWithResponse(ctx, r.provider.SpaceID, *res.JSON201.Id)
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
		response.Diagnostics.AddError("Failed to create volume", fmt.Sprintf("Error waiting for volume (%s) to be created: %s", *res.JSON201.Id, err))
		return
	}

	rr, ok := read.(*api.Volume)
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
	var data resource_volume.VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadVolumesByIdResponse, error) {
		return r.provider.ApiClient.ReadVolumesByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
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
		return r.provider.ApiClient.UpdateVolumeWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString(), body)
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
				return r.provider.ApiClient.ReadVolumesByIdWithResponse(ctx, r.provider.SpaceID, volumeId)
			}, http.StatusOK, &response.Diagnostics)
			if readRes == nil {
				return
			}

			return readRes.JSON200, *readRes.JSON200.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   3 * time.Second,
	}

	time.Sleep(3 * time.Second) // TODO remove when outscale fixes the State field => https://numsproduct.atlassian.net/browse/CLSEXP-612

	read, err := updateStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create volume", fmt.Sprintf("Error waiting for volume (%s) to be created: %s", state.Id.ValueString(), err))
		return
	}

	rr, ok := read.(*api.Volume)
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

func (r *VolumeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_volume.VolumeModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	utils.ExecuteRequest(func() (*api.DeleteVolumeResponse, error) {
		return r.provider.ApiClient.DeleteVolumeWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
}
