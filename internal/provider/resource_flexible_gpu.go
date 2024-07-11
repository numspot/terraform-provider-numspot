package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_flexible_gpu"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &FlexibleGpuResource{}
	_ resource.ResourceWithConfigure   = &FlexibleGpuResource{}
	_ resource.ResourceWithImportState = &FlexibleGpuResource{}
)

type FlexibleGpuResource struct {
	provider Provider
}

func NewFlexibleGpuResource() resource.Resource {
	return &FlexibleGpuResource{}
}

func (r *FlexibleGpuResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *FlexibleGpuResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *FlexibleGpuResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_flexible_gpu"
}

func (r *FlexibleGpuResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_flexible_gpu.FlexibleGpuResourceSchema(ctx)
}

func (r *FlexibleGpuResource) linkVm(ctx context.Context, gpuId string, data resource_flexible_gpu.FlexibleGpuModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Link GPU to VM
	body := LinkFlexibleGpuFromTfToCreateRequest(&data)
	_ = utils.ExecuteRequest(func() (*numspot.LinkFlexibleGpuResponse, error) {
		return r.provider.NumspotClient.LinkFlexibleGpuWithResponse(ctx, r.provider.SpaceID, gpuId, body)
	}, http.StatusNoContent, &diags)
	if diags.HasError() {
		return diags
	}

	// Restart VM needed when linking a GPU
	diags = StopVm(ctx, r.provider, data.VmId.ValueString())
	if diags.HasError() {
		return diags
	}
	diags = StartVm(ctx, r.provider, data.VmId.ValueString())
	if diags.HasError() {
		return diags
	}

	return diags
}

func (r *FlexibleGpuResource) unlinkVm(ctx context.Context, gpuId string, data resource_flexible_gpu.FlexibleGpuModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Unlink GPU from any VM
	_ = utils.ExecuteRequest(func() (*numspot.UnlinkFlexibleGpuResponse, error) {
		return r.provider.NumspotClient.UnlinkFlexibleGpuWithResponse(ctx, r.provider.SpaceID, gpuId)
	}, http.StatusNoContent, &diags)
	if diags.HasError() {
		return diags
	}

	// Restart VM needed when unlinking a GPU
	diags = StopVm(ctx, r.provider, data.VmId.ValueString())
	if diags.HasError() {
		return diags
	}
	diags = StartVm(ctx, r.provider, data.VmId.ValueString())
	if diags.HasError() {
		return diags
	}

	return diags
}

func (r *FlexibleGpuResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		FlexibleGpuFromTfToCreateRequest(&data),
		r.provider.NumspotClient.CreateFlexibleGpuWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Flexible GPU", err.Error())
		return
	}

	createdId := *res.JSON201.Id

	// Link GPU to VM
	if !(data.VmId.IsNull() || data.VmId.IsUnknown()) {
		response.Diagnostics.Append(r.linkVm(ctx, createdId, data)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"attaching", "detaching"},
		[]string{"allocated", "attached"},
		r.provider.NumspotClient.ReadFlexibleGpusByIdWithResponse,
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

func (r *FlexibleGpuResource) read(ctx context.Context, id string, diagnostics diag.Diagnostics) *numspot.FlexibleGpu {
	res, err := r.provider.NumspotClient.ReadFlexibleGpusByIdWithResponse(ctx, r.provider.SpaceID, id)
	if err != nil {
		diagnostics.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		diagnostics.AddError("Failed to read FlexibleGpu", apiError.Error())
		return nil
	}

	return res.JSON200
}

func (r *FlexibleGpuResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	gpu := r.read(ctx, data.Id.ValueString(), response.Diagnostics)
	if gpu == nil {
		return
	}

	tf := FlexibleGpuFromHttpToTf(gpu)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *FlexibleGpuResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if plan.VmId.ValueString() != state.VmId.ValueString() {
		if state.VmId.IsNull() || state.VmId.IsUnknown() { // If GPU is not linked to any VM and we want to link it
			response.Diagnostics.Append(r.linkVm(ctx, state.Id.ValueString(), plan)...)
			if response.Diagnostics.HasError() {
				return
			}

		} else if plan.VmId.IsNull() || plan.VmId.IsUnknown() { // If GPU is linked to a VM and we want to unlink it
			response.Diagnostics.Append(r.unlinkVm(ctx, state.Id.ValueString(), state)...)
			if response.Diagnostics.HasError() {
				return
			}
		} else { // Gpu is linked to a VM and we want to link it to another
			response.Diagnostics.Append(r.unlinkVm(ctx, state.Id.ValueString(), state)...)
			if response.Diagnostics.HasError() {
				return
			}
			response.Diagnostics.Append(r.linkVm(ctx, state.Id.ValueString(), plan)...)
			if response.Diagnostics.HasError() {
				return
			}
		}
	}

	if plan.DeleteOnVmDeletion != state.DeleteOnVmDeletion {
		body := FlexibleGpuFromTfToUpdateRequest(&plan)
		res := utils.ExecuteRequest(func() (*numspot.UpdateFlexibleGpuResponse, error) {
			return r.provider.NumspotClient.UpdateFlexibleGpuWithResponse(
				ctx,
				r.provider.SpaceID,
				state.Id.ValueString(),
				body)
		}, http.StatusOK, &response.Diagnostics)
		if res == nil {
			return
		}

	}

	gpu := r.read(ctx, state.Id.ValueString(), response.Diagnostics)
	if gpu == nil {
		return
	}

	tf := FlexibleGpuFromHttpToTf(gpu)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *FlexibleGpuResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if !(data.VmId.IsNull() || data.VmId.IsUnknown()) {
		response.Diagnostics.Append(r.unlinkVm(ctx, data.Id.ValueString(), data)...)
		// Even if there is an error on unlink, we will try to delete the GPU
	}

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.NumspotClient.DeleteFlexibleGpuWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Flexible GPU", err.Error())
		return
	}
}
