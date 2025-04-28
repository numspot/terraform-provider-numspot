package flexiblegpu

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/flexiblegpu/resource_flexible_gpu"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewFlexibleGpuResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_flexible_gpu"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_flexible_gpu.FlexibleGpuResourceSchema(ctx)
}

func (r *Resource) linkVm(ctx context.Context, gpuId string, data resource_flexible_gpu.FlexibleGpuModel, diags *diag.Diagnostics) {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	// Link GPU to VM
	body := deserializeLinkFlexibleGPU(&data)

	res, err := numspotClient.LinkFlexibleGpuWithResponse(ctx, r.provider.SpaceID, gpuId, body)
	if err != nil {
		diags.AddError("Error while linking Flexible Gpu", err.Error())
		return
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		diags.AddError("Error while parsing Flexible Gpu response", err.Error())
		return
	}
	if diags.HasError() {
		return
	}

	// Restart VM needed when linking a GPU
	//StopVM(ctx, r.provider, data.VmId.ValueString(), diags)
	//if diags.HasError() {
	//	return
	//}
	//vm.StartVm(ctx, r.provider, data.VmId.ValueString(), diags)
	//if diags.HasError() {
	//	return
	//}
}

func (r *Resource) unlinkVm(ctx context.Context, gpuId string, _ resource_flexible_gpu.FlexibleGpuModel, diags *diag.Diagnostics) {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	// Unlink GPU from any VM

	res, err := numspotClient.UnlinkFlexibleGpuWithResponse(ctx, r.provider.SpaceID, gpuId)
	if err != nil {
		diags.AddError("Error while unlinking Flexible Gpu", err.Error())
		return
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		diags.AddError("Error while parsing Flexible Gpu response", err.Error())
		return
	}

	// Restart VM needed when unlinking a GPU
	//vm.StopVm(ctx, r.provider, data.VmId.ValueString(), diags)
	//if diags.HasError() {
	//	return
	//}
	//vm.StartVm(ctx, r.provider, data.VmId.ValueString(), diags)
	//if diags.HasError() {
	//	return
	//}
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		deserializeCreateFlexibleGPU(&data),
		numspotClient.CreateFlexibleGpuWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Flexible GPU", err.Error())
		return
	}

	createdId := *res.JSON201.Id

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

	flexGPU, ok := read.(*api.FlexibleGpu)
	if !ok {
		response.Diagnostics.AddError("Failed to create Flexible GPU", "object conversion error")
		return
	}
	tf := serializeFlexibleGPU(flexGPU)

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *Resource) read(ctx context.Context, id string, diags *diag.Diagnostics) *api.FlexibleGpu {
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

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		diags.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	return res.JSON200
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	gpu := r.read(ctx, data.Id.ValueString(), &response.Diagnostics)
	if gpu == nil {
		return
	}

	tf := serializeFlexibleGPU(gpu)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_flexible_gpu.FlexibleGpuModel
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

	// Handle changes in VM association
	if plan.VmId.ValueString() != state.VmId.ValueString() {
		if state.VmId.IsNull() || state.VmId.IsUnknown() { // If GPU is not linked to any VM, we want to link it
			r.linkVm(ctx, state.Id.ValueString(), plan, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}

		} else if plan.VmId.IsNull() || plan.VmId.IsUnknown() { // If GPU is linked to a VM, we want to unlink it
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
		} else { // Gpu is linked to a VM, we want to link it to another
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

	// Update delete_on_vm_deletion flag if changed
	if plan.DeleteOnVmDeletion != state.DeleteOnVmDeletion {
		body := deserializeUpdateFlexibleGPU(&plan)

		res, err := numspotClient.UpdateFlexibleGpuWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString(), body)
		if err != nil {
			response.Diagnostics.AddError("unable to update flexible gpu", err.Error())
			return
		}

		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			response.Diagnostics.AddError("unable to update flexible gpu", err.Error())
			return
		}
	}

	gpu := r.read(ctx, state.Id.ValueString(), &response.Diagnostics)
	if gpu == nil {
		return
	}

	tf := serializeFlexibleGPU(gpu)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	// Unlink GPU from VM if it's attached
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

func serializeFlexibleGPU(http *api.FlexibleGpu) resource_flexible_gpu.FlexibleGpuModel {
	return resource_flexible_gpu.FlexibleGpuModel{
		DeleteOnVmDeletion:   types.BoolPointerValue(http.DeleteOnVmDeletion),
		Generation:           types.StringPointerValue(http.Generation),
		Id:                   types.StringPointerValue(http.Id),
		ModelName:            types.StringPointerValue(http.ModelName),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringValue(utils.ConvertAzNamePtrToString(http.AvailabilityZoneName)),
		VmId:                 types.StringPointerValue(http.VmId),
	}
}

func deserializeLinkFlexibleGPU(tf *resource_flexible_gpu.FlexibleGpuModel) api.LinkFlexibleGpuJSONRequestBody {
	vmId := utils.FromTfStringToStringPtr(tf.VmId)
	return api.LinkFlexibleGpuJSONRequestBody{
		VmId: utils.GetPtrValue(vmId),
	}
}

func deserializeCreateFlexibleGPU(tf *resource_flexible_gpu.FlexibleGpuModel) api.CreateFlexibleGpuJSONRequestBody {
	return api.CreateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion:   tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generation:           tf.Generation.ValueStringPointer(),
		ModelName:            tf.ModelName.ValueString(),
		AvailabilityZoneName: api.AvailabilityZoneName(tf.AvailabilityZoneName.ValueString()),
	}
}

func deserializeUpdateFlexibleGPU(tf *resource_flexible_gpu.FlexibleGpuModel) api.UpdateFlexibleGpuJSONRequestBody {
	return api.UpdateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.DeleteOnVmDeletion),
	}
}
