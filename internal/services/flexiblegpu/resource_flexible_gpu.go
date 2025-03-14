package flexiblegpu

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
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

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_flexible_gpu"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = FlexibleGpuResourceSchema(ctx)
}

func (r *Resource) linkVm(ctx context.Context, gpuId string, data FlexibleGpuModel, diags *diag.Diagnostics) {
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

func (r *Resource) unlinkVm(ctx context.Context, gpuId string, _ FlexibleGpuModel, diags *diag.Diagnostics) {
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
		deserializeCreateFlexibleGPU(&data),
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
	tf := serializeFlexibleGPU(flexGPU)

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *Resource) read(ctx context.Context, id string, diags *diag.Diagnostics) *numspot.FlexibleGpu {
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
	var data FlexibleGpuModel
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

func serializeFlexibleGPU(http *numspot.FlexibleGpu) FlexibleGpuModel {
	return FlexibleGpuModel{
		DeleteOnVmDeletion:   types.BoolPointerValue(http.DeleteOnVmDeletion),
		Generation:           types.StringPointerValue(http.Generation),
		Id:                   types.StringPointerValue(http.Id),
		ModelName:            types.StringPointerValue(http.ModelName),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		VmId:                 types.StringPointerValue(http.VmId),
	}
}

func deserializeLinkFlexibleGPU(tf *FlexibleGpuModel) numspot.LinkFlexibleGpuJSONRequestBody {
	vmId := utils.FromTfStringToStringPtr(tf.VmId)
	return numspot.LinkFlexibleGpuJSONRequestBody{
		VmId: utils.GetPtrValue(vmId),
	}
}

func deserializeCreateFlexibleGPU(tf *FlexibleGpuModel) numspot.CreateFlexibleGpuJSONRequestBody {
	return numspot.CreateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion:   tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generation:           tf.Generation.ValueStringPointer(),
		ModelName:            tf.ModelName.ValueString(),
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueString(),
	}
}

func deserializeUpdateFlexibleGPU(tf *FlexibleGpuModel) numspot.UpdateFlexibleGpuJSONRequestBody {
	return numspot.UpdateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.DeleteOnVmDeletion),
	}
}
