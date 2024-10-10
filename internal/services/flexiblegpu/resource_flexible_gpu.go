package flexiblegpu

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &FlexibleGpuResource{}
	_ resource.ResourceWithConfigure   = &FlexibleGpuResource{}
	_ resource.ResourceWithImportState = &FlexibleGpuResource{}
)

type FlexibleGpuResource struct {
	provider *client.NumSpotSDK
}

func NewFlexibleGpuResource() resource.Resource {
	return &FlexibleGpuResource{}
}

func (r *FlexibleGpuResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *FlexibleGpuResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *FlexibleGpuResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_flexible_gpu"
}

func (r *FlexibleGpuResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = FlexibleGpuResourceSchema(ctx)
}

func (r *FlexibleGpuResource) linkVm(ctx context.Context, gpuId string, data FlexibleGpuModel, diags *diag.Diagnostics) {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	// Link GPU to VM
	body := LinkFlexibleGpuFromTfToCreateRequest(&data)
	_ = utils.ExecuteRequest(func() (*numspot.LinkFlexibleGpuResponse, error) {
		return numspotClient.LinkFlexibleGpuWithResponse(ctx, r.provider.SpaceID, gpuId, body)
	}, http.StatusNoContent, diags)
	if diags.HasError() {
		return
	}

	// Restart VM needed when linking a GPU
	vm.StopVm(ctx, r.provider, data.VmId.ValueString(), diags)
	if diags.HasError() {
		return
	}
	vm.StartVm(ctx, r.provider, data.VmId.ValueString(), diags)
	if diags.HasError() {
		return
	}
}

func (r *FlexibleGpuResource) unlinkVm(ctx context.Context, gpuId string, data FlexibleGpuModel, diags *diag.Diagnostics) {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	// Unlink GPU from any VM
	_ = utils.ExecuteRequest(func() (*numspot.UnlinkFlexibleGpuResponse, error) {
		return numspotClient.UnlinkFlexibleGpuWithResponse(ctx, r.provider.SpaceID, gpuId)
	}, http.StatusNoContent, diags)
	if diags.HasError() {
		return
	}

	// Restart VM needed when unlinking a GPU
	vm.StopVm(ctx, r.provider, data.VmId.ValueString(), diags)
	if diags.HasError() {
		return
	}
	vm.StartVm(ctx, r.provider, data.VmId.ValueString(), diags)
	if diags.HasError() {
		return
	}
}

func (r *FlexibleGpuResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data FlexibleGpuModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		FlexibleGpuFromTfToCreateRequest(&data),
		numspotClient.CreateFlexibleGpuWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Flexible GPU", err.Error())
		return
	}

	createdId := *res.JSON201.Id

	// Link GPU to VM
	if !(data.VmId.IsNull() || data.VmId.IsUnknown()) {
		r.linkVm(ctx, createdId, data, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	read, err := utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"attaching", "detaching"},
		[]string{"allocated", "attached"},
		numspotClient.ReadFlexibleGpusByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Flexible GPU", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", data.Id.ValueString(), err))
		return
	}

	flexGPU, ok := read.(*numspot.FlexibleGpu)
	if !ok {
		response.Diagnostics.AddError("Failed to create Flexible GPU", "object conversion error")
		return
	}
	tf := FlexibleGpuFromHttpToTf(flexGPU)

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *FlexibleGpuResource) read(ctx context.Context, id string, diags *diag.Diagnostics) *numspot.FlexibleGpu {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}
	res, err := numspotClient.ReadFlexibleGpusByIdWithResponse(ctx, r.provider.SpaceID, id)
	if err != nil {
		diags.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		diags.AddError("Failed to read FlexibleGpu", apiError.Error())
		return nil
	}

	return res.JSON200
}

func (r *FlexibleGpuResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	gpu := r.read(ctx, data.Id.ValueString(), &response.Diagnostics)
	if gpu == nil {
		return
	}

	tf := FlexibleGpuFromHttpToTf(gpu)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *FlexibleGpuResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state FlexibleGpuModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	if plan.VmId.ValueString() != state.VmId.ValueString() {
		if state.VmId.IsNull() || state.VmId.IsUnknown() { // If GPU is not linked to any VM and we want to link it
			r.linkVm(ctx, state.Id.ValueString(), plan, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}

		} else if plan.VmId.IsNull() || plan.VmId.IsUnknown() { // If GPU is linked to a VM and we want to unlink it
			var diagnostics diag.Diagnostics // Use a temporary diag because some errors might be ok here
			r.unlinkVm(ctx, state.Id.ValueString(), state, &diagnostics)
			if diagnostics.HasError() {
				_, err = utils.RetryReadUntilStateValid(
					ctx,
					state.Id.ValueString(),
					r.provider.SpaceID,
					[]string{"detaching"},
					[]string{"allocated"},
					numspotClient.ReadFlexibleGpusByIdWithResponse,
				)
				if err != nil {
					response.Diagnostics.Append(diagnostics...)
					response.Diagnostics.AddError("Failed while waiting for GPU to get unlinked", err.Error())
				}
			}
		} else { // Gpu is linked to a VM and we want to link it to another
			var diagnostics diag.Diagnostics // Use a temporary diag because some errors might be ok here
			r.unlinkVm(ctx, state.Id.ValueString(), state, &diagnostics)
			if diagnostics.HasError() {
				_, err = utils.RetryReadUntilStateValid(
					ctx,
					state.Id.ValueString(),
					r.provider.SpaceID,
					[]string{"detaching"},
					[]string{"allocated"},
					numspotClient.ReadFlexibleGpusByIdWithResponse,
				)
				if err != nil {
					response.Diagnostics.Append(diagnostics...)
					response.Diagnostics.AddError("Failed while waiting for GPU to get unlinked", err.Error())
				}
			}
			r.linkVm(ctx, state.Id.ValueString(), plan, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}
		}
	}

	if plan.DeleteOnVmDeletion != state.DeleteOnVmDeletion {
		body := FlexibleGpuFromTfToUpdateRequest(&plan)
		res := utils.ExecuteRequest(func() (*numspot.UpdateFlexibleGpuResponse, error) {
			return numspotClient.UpdateFlexibleGpuWithResponse(
				ctx,
				r.provider.SpaceID,
				state.Id.ValueString(),
				body)
		}, http.StatusOK, &response.Diagnostics)
		if res == nil {
			return
		}

	}

	gpu := r.read(ctx, state.Id.ValueString(), &response.Diagnostics)
	if gpu == nil {
		return
	}

	tf := FlexibleGpuFromHttpToTf(gpu)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *FlexibleGpuResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	if !(data.VmId.IsNull() || data.VmId.IsUnknown()) {
		var diagnostics diag.Diagnostics // Use a temporary diag because some errors might be ok here
		r.unlinkVm(ctx, data.Id.ValueString(), data, &diagnostics)
		if diagnostics.HasError() {
			_, err = utils.RetryReadUntilStateValid(
				ctx,
				data.Id.ValueString(),
				r.provider.SpaceID,
				[]string{"detaching"},
				[]string{"allocated"},
				numspotClient.ReadFlexibleGpusByIdWithResponse,
			)
			if err != nil {
				response.Diagnostics.Append(diagnostics...)
				response.Diagnostics.AddError("Failed while waiting for GPU to get unlinked", err.Error())
				return
			}
		}
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), numspotClient.DeleteFlexibleGpuWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Flexible GPU", err.Error())
		return
	}
}
