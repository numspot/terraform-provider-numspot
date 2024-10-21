package vm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VmResource{}
	_ resource.ResourceWithConfigure   = &VmResource{}
	_ resource.ResourceWithImportState = &VmResource{}
)

type VmResource struct {
	provider *client.NumSpotSDK
}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

func (r *VmResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *VmResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VmResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vm"
}

func (r *VmResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VmResourceSchema(ctx)
}

func (r *VmResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data VmModel
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
		VmFromTfToCreateRequest(ctx, &data, &response.Diagnostics),
		numspotClient.CreateVmsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", err.Error())
	}
	if response.Diagnostics.HasError() {
		return
	}

	vm := *res.JSON201
	createdId := *vm.Id

	// Create tags
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, numspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	read, err := utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending", "stopping"},
		[]string{"running", "stopped"}, // In some cases, when there is insufficient capacity the VM is created with state = stopped
		numspotClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", createdId, err))
		return
	}

	vmSchema, ok := read.(*numspot.Vm)
	if !ok {
		response.Diagnostics.AddError("Failed to create VM", "object conversion error")
		return
	}

	// In some cases, when there is insufficient capacity the VM is created with state = stopped
	if utils.GetPtrValue(vmSchema.State) == "stopped" {
		response.Diagnostics.AddError("Issue while creating VM", fmt.Sprintf("VM was created in 'stopped' state. Reason : %s", utils.GetPtrValue(vmSchema.StateReason)))
		return
	}

	tf := VmFromHttpToTf(ctx, vmSchema, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VmResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadVmsByIdResponse, error) {
		id := utils.FromTfStringToStringPtr(data.Id)
		if id == nil {
			return nil, errors.New("Found invalid id")
		}
		return numspotClient.ReadVmsByIdWithResponse(ctx, r.provider.SpaceID, *id)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := VmFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VmResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VmModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	vmId := state.Id.ValueString()

	// Update tags
	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			numspotClient,
			r.provider.SpaceID,
			vmId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	body := VmFromTfToUpdaterequest(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	bodyFromState := VmFromTfToUpdaterequest(ctx, &state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if isUpdateNeeded(body, bodyFromState) {
		// Stop VM before doing update
		StopVm(ctx, r.provider, vmId, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		// Update VM
		updatedRes := utils.ExecuteRequest(func() (*numspot.UpdateVmResponse, error) {
			return numspotClient.UpdateVmWithResponse(ctx, r.provider.SpaceID, vmId, body)
		}, http.StatusOK, &response.Diagnostics)

		if updatedRes == nil || response.Diagnostics.HasError() {
			return
		}

		// Restart VM
		StartVm(ctx, r.provider, vmId, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Retries read on VM until state is OK
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		vmId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"running"},
		numspotClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to update VM", fmt.Sprintf("Error waiting for VM to be created: %s", err))
		return
	}
	vmObject, ok := read.(*numspot.Vm)
	if !ok {
		response.Diagnostics.AddError("Failed to update VM", "object conversion error")
		return
	}

	tf := VmFromHttpToTf(ctx, vmObject, &response.Diagnostics)

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func compareSimpleFieldPtr[R comparable](val1 *R, val2 *R) bool {
	return utils.GetPtrValue(val1) == utils.GetPtrValue(val2)
}

func compareSlicePtr[R comparable](val1 *[]R, val2 *[]R) bool {
	return slices.Equal(utils.GetPtrValue(val1), utils.GetPtrValue(val2))
}

func isUpdateNeeded(plan numspot.UpdateVmJSONRequestBody, state numspot.UpdateVmJSONRequestBody) bool {
	return !(compareSimpleFieldPtr(plan.BsuOptimized, state.BsuOptimized) &&
		compareSimpleFieldPtr(plan.DeletionProtection, state.DeletionProtection) &&
		compareSimpleFieldPtr(plan.KeypairName, state.KeypairName) &&
		compareSimpleFieldPtr(plan.NestedVirtualization, state.NestedVirtualization) &&
		(utils.GetPtrValue(plan.Performance) == "" || compareSimpleFieldPtr(plan.Performance, state.Performance)) && // if performance is not provided by user,
		(len(utils.GetPtrValue(plan.BlockDeviceMappings)) == 0) &&
		compareSimpleFieldPtr(plan.UserData, state.UserData) &&
		compareSimpleFieldPtr(plan.VmInitiatedShutdownBehavior, state.VmInitiatedShutdownBehavior) &&
		compareSimpleFieldPtr(plan.Type, state.Type) &&
		(len(utils.GetPtrValue(plan.BlockDeviceMappings)) == 0 || (compareSlicePtr(plan.BlockDeviceMappings, state.BlockDeviceMappings))) &&
		compareSimpleFieldPtr(plan.IsSourceDestChecked, state.IsSourceDestChecked))
}

func (r *VmResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	err = utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), numspotClient.DeleteVmsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VM", err.Error())
		return
	}
}
