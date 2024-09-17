package volume

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
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

	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
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

	// Wait for volume to be ready before linking to VM
	read, err := utils.RetryReadUntilStateValid(
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

	// Link Volume to a VM
	if linkVmConfigured(data) {
		utils.ExecuteRequest(func() (*numspot.LinkVolumeResponse, error) {
			return r.provider.GetNumspotClient().LinkVolumeWithResponse(
				ctx,
				r.provider.GetSpaceID(),
				createdId,
				VolumeFromTfToLinkRequest(&data),
			)
		}, http.StatusNoContent, &response.Diagnostics)

		if response.Diagnostics.HasError() {
			return
		}

		// Wait for volume to be linked to VM
		createStateConf := &retry.StateChangeConf{
			Pending: []string{"attaching"},
			Target:  []string{"attached"},
			Refresh: func() (interface{}, string, error) {
				readRes, err := r.provider.GetNumspotClient().ReadVolumesByIdWithResponse(ctx, r.provider.GetSpaceID(), createdId)
				if err != nil {
					return nil, "", fmt.Errorf("failed to read volume : %v", err)
				}

				if len(*readRes.JSON200.LinkedVolumes) > 0 {
					linkState := (*readRes.JSON200.LinkedVolumes)[0].State
					return readRes.JSON200, *linkState, nil
				}

				return nil, "", fmt.Errorf("Volume not linked to any VM : %v", err)
			},
			Timeout: utils.TfRequestRetryTimeout,
			Delay:   utils.TfRequestRetryDelay,
		}
		read, err = createStateConf.WaitForStateContext(ctx)
		if err != nil {
			response.Diagnostics.AddError("Failed to create volume", fmt.Sprintf("Error waiting for volume (%s) to be linked: %s", *res.JSON201.Id, err))
			return
		}
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

	res := utils.ExecuteRequest(func() (*numspot.ReadVolumesByIdResponse, error) {
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

func shouldUpdate(plan, state VolumeModel) bool {
	shouldUpdate := false
	shouldUpdate = shouldUpdate || (!utils.IsTfValueNull(plan.Size) && !plan.Size.Equal(state.Size))
	shouldUpdate = shouldUpdate || (!utils.IsTfValueNull(plan.Type) && !plan.Type.Equal(state.Type))
	shouldUpdate = shouldUpdate || (!utils.IsTfValueNull(plan.Iops) && !plan.Iops.Equal(state.Iops))

	return shouldUpdate
}

func (r *VolumeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VolumeModel
	var diags diag.Diagnostics
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

	stateVmExists := true
	if linkVmConfigured(state) {
		// Stop linked VM before unlinking Volume
		diags, stateVmExists = r.stopVm(ctx, state.LinkVM.VmID.ValueString())
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	}

	if shouldUpdate(plan, state) {
		updatedRes := utils.ExecuteRequest(func() (*numspot.UpdateVolumeResponse, error) {
			body := ValueFromTfToUpdaterequest(&plan)
			return r.provider.GetNumspotClient().UpdateVolumeWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString(), body)
		}, http.StatusOK, &response.Diagnostics)
		if updatedRes == nil {
			return
		}
	}

	time.Sleep(3 * time.Second) // TODO remove when outscale fixes the State field => https://numsproduct.atlassian.net/browse/CLSEXP-612

	// Update Link to VM
	if !state.LinkVM.Equal(plan.LinkVM) {
		var diags diag.Diagnostics
		if linkVmConfigured(state) && stateVmExists {
			utils.ExecuteRequest(func() (*numspot.UnlinkVolumeResponse, error) {
				return r.provider.GetNumspotClient().UnlinkVolumeWithResponse(
					ctx,
					r.provider.GetSpaceID(),
					state.Id.ValueString(),
					VolumeFromTfToUnlinkRequest(&state),
				)
			}, http.StatusNoContent, &diags)
		}

		if linkVmConfigured(plan) {
			utils.ExecuteRequest(func() (*numspot.LinkVolumeResponse, error) {
				return r.provider.GetNumspotClient().LinkVolumeWithResponse(
					ctx,
					r.provider.GetSpaceID(),
					state.Id.ValueString(),
					VolumeFromTfToLinkRequest(&plan),
				)
			}, http.StatusNoContent, &response.Diagnostics)
		}

		if response.Diagnostics.HasError() {
			return
		}
	}

	if linkVmConfigured(plan) {
		// Start linked VM after updating Volume
		diags := vm.StartVm(ctx, r.provider, plan.LinkVM.VmID.ValueString())
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	}

	volumeId := state.Id.ValueString()
	// Retries read on resource until state is OK
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		volumeId,
		r.provider.GetSpaceID(),
		[]string{"creating", "updating"},
		[]string{"available", "in-use"},
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

func (r *VolumeResource) stopVm(ctx context.Context, vmId string) (diag.Diagnostics, bool) {
	diags := vm.StopVm(ctx, r.provider, vmId)
	if diags.HasError() {
		if !vm.VmIsDeleted(vmId) {
			return diags, true
		} else { // Else, the VM got removed, which is OK
			return nil, false
		}
	}

	return nil, true
}

func linkVmConfigured(data VolumeModel) bool {
	return !utils.IsTfValueNull(data.LinkVM) && !utils.IsTfValueNull(data.LinkVM.VmID)
}

func (r *VolumeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VolumeModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if linkVmConfigured(data) {
		// Stop linked VM before unlinking Volume
		diags, stateVmExists := r.stopVm(ctx, data.LinkVM.VmID.ValueString())

		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		if stateVmExists {
			// Unlink VM
			utils.ExecuteRequest(func() (*numspot.UnlinkVolumeResponse, error) {
				return r.provider.GetNumspotClient().UnlinkVolumeWithResponse(
					ctx,
					r.provider.GetSpaceID(),
					data.Id.ValueString(),
					VolumeFromTfToUnlinkRequest(&data),
				)
			}, http.StatusNoContent, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}
			diags = vm.StartVm(ctx, r.provider, data.LinkVM.VmID.ValueString())
			response.Diagnostics.Append(diags...)

			if response.Diagnostics.HasError() {
				return
			}

		}
	}

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteVolumeWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Volume", err.Error())
		return
	}
}
