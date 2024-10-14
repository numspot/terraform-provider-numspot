package core

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	vmPendingStates = []string{pending, stopping, shuttingDown}
	vmTargetStates  = []string{running, stopped, shutdown, terminated}
)

func CreateVM(ctx context.Context, provider *client.NumSpotSDK, numSpotVMCreate numspot.CreateVmsJSONRequestBody, tags []numspot.ResourceTag) (numSpotVM *numspot.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreateVmsResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotVMCreate,
		numspotClient.CreateVmsWithResponse); err != nil {
		return nil, err
	}

	vmID := *retryCreate.JSON201.Id

	//// Create tags
	//if len(data.Tags.Elements()) > 0 {
	//	tags.CreateTagsFromTf(ctx, numspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//}
	//
	//read, err := utils.RetryReadUntilStateValid(
	//	ctx,
	//	createdId,
	//	r.provider.SpaceID,
	//	[]string{"pending"},
	//	[]string{"running", "stopped"}, // In some cases, when there is insufficient capacity the VM is created with state = stopped
	//	numspotClient.ReadVmsByIdWithResponse,
	//)
	//if err != nil {
	//	response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", createdId, err))
	//	return
	//}
	//
	//vmSchema, ok := read.(*numspot.Vm)
	//if !ok {
	//	response.Diagnostics.AddError("Failed to create VM", "object conversion error")
	//	return
	//}
	//
	//// In some cases, when there is insufficient capacity the VM is created with state = stopped
	//if utils.GetPtrValue(vmSchema.State) == "stopped" {
	//	response.Diagnostics.AddError("Issue while creating VM", fmt.Sprintf("VM was created in 'stopped' state. Reason : %s", utils.GetPtrValue(vmSchema.StateReason)))
	//	return
	//}
	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, vmID, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

func UpdateVMAttributes(ctx context.Context, provider *client.NumSpotSDK, numSpotVMCreate numspot.UpdateVmJSONRequestBody, vmID string) (numSpotVM *numspot.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	//body := VmFromTfToUpdaterequest(ctx, &plan, &response.Diagnostics)
	//if response.Diagnostics.HasError() {
	//	return
	//}
	//bodyFromState := VmFromTfToUpdaterequest(ctx, &state, &response.Diagnostics)
	//if response.Diagnostics.HasError() {
	//	return
	//}
	//
	//if isUpdateNeeded(body, bodyFromState) {
	//	// Stop VM before doing update
	StopVM(ctx, r.provider, vmId, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	//
	//	// Update VM
	//	updatedRes := utils.ExecuteRequest(func() (*numspot.UpdateVmResponse, error) {
	var updateVMResponse *numspot.UpdateVmResponse
	updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, provider.SpaceID, vmID, body)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(updateVMResponse.Body, updateVMResponse.StatusCode()); err != nil {
		return nil, err
	}

	//}, http.StatusOK, &response.Diagnostics)
	//
	//if updatedRes == nil || response.Diagnostics.HasError() {
	//	return
	//}

	// Restart VM
	StartVM(ctx, r.provider, vmId, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries read on VM until state is OK
	//read, err := utils.RetryReadUntilStateValid(
	//	ctx,
	//	vmId,
	//	r.provider.SpaceID,
	//	[]string{"pending"},
	//	[]string{"running"},
	//	numspotClient.ReadVmsByIdWithResponse,
	//)
	//if err != nil {
	//	response.Diagnostics.AddError("Failed to update VM", fmt.Sprintf("Error waiting for VM to be created: %s", err))
	//	return
	//}

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

	//vmObject, ok := read.(*numspot.Vm)
	//if !ok {
	//	response.Diagnostics.AddError("Failed to update VM", "object conversion error")
	//	return
	//}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

func UpdateVMTags(ctx context.Context, provider *client.NumSpotSDK, numSpotVMCreate numspot.UpdateVmJSONRequestBody, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (numSpotVM *numspot.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	//body := VmFromTfToUpdaterequest(ctx, &plan, &response.Diagnostics)
	//if response.Diagnostics.HasError() {
	//	return
	//}

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
	//read, err := utils.RetryReadUntilStateValid(
	//	ctx,
	//	vmId,
	//	r.provider.SpaceID,
	//	[]string{"pending"},
	//	[]string{"running"},
	//	numspotClient.ReadVmsByIdWithResponse,
	//)
	//if err != nil {
	//	response.Diagnostics.AddError("Failed to update VM", fmt.Sprintf("Error waiting for VM to be created: %s", err))
	//	return
	//}

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

	//vmObject, ok := read.(*numspot.Vm)
	//if !ok {
	//	response.Diagnostics.AddError("Failed to update VM", "object conversion error")
	//	return
	//}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

func RetryReadVM(ctx context.Context, provider *client.NumSpotSDK, op string, vmID string) (*numspot.Vm, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, vmID, provider.SpaceID, vmPendingStates, vmTargetStates, numspotClient.ReadVmsByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotVM, assert := read.(*numspot.Vm)
	if !assert {
		return nil, fmt.Errorf("invalid VM assertion %s: %s", vmID, op)
	}
	return numSpotVM, err
}

func ReadVM(ctx context.Context, provider *client.NumSpotSDK, vmID string) (*numspot.Vm, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, vmID, provider.SpaceID, vmPendingStates, vmTargetStates, numspotClient.ReadVmsByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotVM, assert := read.(*numspot.Vm)
	if !assert {
		return nil, fmt.Errorf("invalid VM assertion %s: %s", vmID, op)
	}
	return numSpotVM, err
}

func StopVM(ctx context.Context, provider *client.NumSpotSDK, vm string) (err error) {
	// Already stopped
	if vm == "" {
		return nil
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	var vmStatus *numspot.ReadVmsByIdResponse
	if vmStatus, err = numspotClient.ReadVmsByIdWithResponse(ctx, provider.SpaceID, vm); err != nil {
		return err
	}

	// VM does not exist
	if vmStatus == nil || vmStatus.JSON200 == nil {
		return nil
	}
	if *vmStatus.JSON200.State == "stopped" || *vmStatus.JSON200.State == "terminated" {
		return nil
	}

	//////////////////
	forceStop := true
	// Stop the VM
	if _, err = numspotClient.StopVmWithResponse(ctx, provider.SpaceID, vm, numspot.StopVm{ForceStop: &forceStop}); err != nil {
		return err
	}

	if _, err = utils.RetryReadUntilStateValid(ctx, vm, provider.SpaceID, []string{"stopping"}, []string{"stopped", "terminated"},
		numspotClient.ReadVmsByIdWithResponse); err != nil {
		return err
	}

	return nil
}

func StartVM(ctx context.Context, provider *client.NumSpotSDK, vm string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	// Already running
	if vm == "" {
		return nil
	}

	var vmStatus *numspot.ReadVmsByIdResponse
	if vmStatus, err = numspotClient.ReadVmsByIdWithResponse(
		ctx,
		provider.SpaceID,
		vm,
	); err != nil {
		return err
	}

	// VM does not exist
	if vmStatus == nil || vmStatus.JSON200 == nil {
		return nil
	}

	if *vmStatus.JSON200.State == "running" || *vmStatus.JSON200.State == "terminated" {
		return nil
	}

	//////////////////
	// Start the VM
	if _, err = numspotClient.StartVmWithResponse(ctx, provider.SpaceID, vm); err != nil {
		return err
	}

	_, err = utils.RetryReadUntilStateValid(
		ctx,
		vm,
		provider.SpaceID,
		[]string{"pending"},
		[]string{"running"},
		numspotClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		return err
	}

	return nil
}
