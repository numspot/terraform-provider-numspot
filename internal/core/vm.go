package core

import (
	"context"
	"fmt"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
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

	if len(tags) > 0 {
		if err = createTags(ctx, provider, vmID, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

func UpdateVMAttributes(ctx context.Context, provider *client.NumSpotSDK, numSpotVMUpdate numspot.UpdateVmJSONRequestBody, vmID string) (numSpotVM *numspot.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	if err = StopVM(ctx, provider, vmID); err != nil {
		return nil, err
	}

	numSpotVMUpdate.KeypairName = nil

	var updateVMResponse *numspot.UpdateVmResponse
	if updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, spaceID, vmID, numSpotVMUpdate); err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(updateVMResponse.Body, updateVMResponse.StatusCode()); err != nil {
		return nil, err
	}

	if err = StartVM(ctx, provider, vmID); err != nil {
		return nil, err
	}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

func UpdateVMKeypair(ctx context.Context, provider *client.NumSpotSDK, numSpotVMUpdate numspot.UpdateVmJSONRequestBody, vmID string) (numSpotVM *numspot.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	if err = StopVM(ctx, provider, vmID); err != nil {
		return nil, err
	}

	numSpotVMUpdate.BlockDeviceMappings = nil
	numSpotVMUpdate.BsuOptimized = nil
	numSpotVMUpdate.IsSourceDestChecked = nil
	numSpotVMUpdate.SecurityGroupIds = nil
	numSpotVMUpdate.DeletionProtection = nil
	numSpotVMUpdate.UserData = nil
	numSpotVMUpdate.Type = nil
	numSpotVMUpdate.SecurityGroupIds = nil
	numSpotVMUpdate.VmInitiatedShutdownBehavior = nil
	numSpotVMUpdate.NestedVirtualization = nil

	var updateVMResponse *numspot.UpdateVmResponse
	if updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, spaceID, vmID, numSpotVMUpdate); err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(updateVMResponse.Body, updateVMResponse.StatusCode()); err != nil {
		return nil, err
	}

	if err = StartVM(ctx, provider, vmID); err != nil {
		return nil, err
	}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

func UpdateVMTags(ctx context.Context, provider *client.NumSpotSDK, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, vmID string) (numSpotVM *numspot.Vm, err error) {
	if err = updateResourceTags(ctx, provider, stateTags, planTags, vmID); err != nil {
		return nil, err
	}
	return RetryReadVM(ctx, provider, updateOp, vmID)
}

func DeleteVM(ctx context.Context, provider *client.NumSpotSDK, vmID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, vmID, numspotClient.DeleteVmsWithResponse)
	if err != nil {
		return err
	}

	return nil
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
		return nil, fmt.Errorf("invalid vm assertion %s: %s", vmID, op)
	}
	return numSpotVM, err
}

func ReadVM(ctx context.Context, provider *client.NumSpotSDK, vmID string) (*numspot.Vm, error) {
	var numSpotReadVM *numspot.ReadVmsByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadVM, err = numspotClient.ReadVmsByIdWithResponse(ctx, provider.SpaceID, vmID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadVM.Body, numSpotReadVM.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadVM.JSON200, err
}

func ReadVMsWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadVmsParams) (*[]numspot.Vm, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadVM, err := numspotClient.ReadVmsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadVM.Body, numSpotReadVM.StatusCode()); err != nil {
		return nil, err
	}

	if numSpotReadVM.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of vms but got nil")
	}

	return numSpotReadVM.JSON200.Items, err
}

func StopVM(ctx context.Context, provider *client.NumSpotSDK, vm string) (err error) {
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
	if *vmStatus.JSON200.State == stopped || *vmStatus.JSON200.State == terminated {
		return nil
	}

	//////////////////
	forceStop := true
	// Stop the VM
	if _, err = numspotClient.StopVmWithResponse(ctx, provider.SpaceID, vm, numspot.StopVm{ForceStop: &forceStop}); err != nil {
		return err
	}

	if _, err = utils.RetryReadUntilStateValid(ctx, vm, provider.SpaceID, []string{stopping}, []string{stopped, terminated},
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
	if vmStatus, err = numspotClient.ReadVmsByIdWithResponse(ctx, provider.SpaceID, vm); err != nil {
		return err
	}
	// VM does not exist
	if vmStatus == nil || vmStatus.JSON200 == nil {
		return nil
	}
	if *vmStatus.JSON200.State == running || *vmStatus.JSON200.State == terminated {
		return nil
	}

	//////////////////
	// Start the VM
	if _, err = numspotClient.StartVmWithResponse(ctx, provider.SpaceID, vm); err != nil {
		return err
	}

	if _, err = utils.RetryReadUntilStateValid(ctx, vm, provider.SpaceID, []string{pending}, []string{running}, numspotClient.ReadVmsByIdWithResponse); err != nil {
		return err
	}

	return nil
}
