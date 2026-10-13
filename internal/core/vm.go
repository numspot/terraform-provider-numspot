package core

import (
	"context"
	"errors"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	vmPendingStates = []string{pending, stopping, shuttingDown}
	vmTargetStates  = []string{running, stopped, shutdown, terminated}

	// errInvalidVMAssertion is returned when a VM read result cannot be asserted
	// to the expected type.
	errInvalidVMAssertion = errors.New("invalid vm assertion")
	// errExpectedVMList is returned when a VM list response is unexpectedly nil.
	errExpectedVMList = errors.New("HTTP call failed : expected a list of vms but got nil")
)

// CreateVM creates VM.
func CreateVM(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotVMCreate *api.CreateVmsJSONRequestBody,
	tags []api.ResourceTag,
) (numSpotVM *api.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	// BootOnCreation is not a Terraform attribute; the resource always boots the
	// VM on creation, so set it here rather than in the (generated) request body.
	if numSpotVMCreate.BootOnCreation == nil {
		numSpotVMCreate.BootOnCreation = utils.PointerOf(true)
	}

	var retryCreate *api.CreateVmsResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, *numSpotVMCreate,
		numspotClient.CreateVmsWithResponse); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	vmID := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, vmID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

// UpdateVMAttributes updates VM attributes.
func UpdateVMAttributes(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotVMUpdate *api.UpdateVmJSONRequestBody,
	vmID string,
) (numSpotVM *api.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err = StopVM(ctx, provider, vmID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numSpotVMUpdate.KeypairName = nil

	var updateVMResponse *api.UpdateVmResponse
	if updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, spaceID, vmID, *numSpotVMUpdate); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	if err = utils.ParseHTTPError(updateVMResponse.Body, updateVMResponse.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err = StartVM(ctx, provider, vmID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

// UpdateVMSecurityGroups updates the security groups of a running VM without
// stopping it. The API applies a security group change to a running VM
// (verified on preprod), so a SG-only change must not force the
// stop/update/start cycle of UpdateVMAttributes (CLOS-3261).
func UpdateVMSecurityGroups(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotVMUpdate *api.UpdateVmJSONRequestBody,
	vmID string,
) (numSpotVM *api.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numSpotVMUpdate.KeypairName = nil
	numSpotVMUpdate.BlockDeviceMappings = nil
	numSpotVMUpdate.DeletionProtection = nil
	numSpotVMUpdate.NestedVirtualization = nil
	numSpotVMUpdate.Type = nil
	numSpotVMUpdate.UserData = nil
	numSpotVMUpdate.VmInitiatedShutdownBehavior = nil
	numSpotVMUpdate.IsSourceDestChecked = nil

	var updateVMResponse *api.UpdateVmResponse
	if updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, spaceID, vmID, *numSpotVMUpdate); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	if err = utils.ParseHTTPError(updateVMResponse.Body, updateVMResponse.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return RetryReadVM(ctx, provider, updateOp, vmID)
}

// UpdateVMKeypair updates VM keypair.
func UpdateVMKeypair(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotVMUpdate *api.UpdateVmJSONRequestBody,
	vmID string,
) (numSpotVM *api.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err = StopVM(ctx, provider, vmID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numSpotVMUpdate.BlockDeviceMappings = nil
	numSpotVMUpdate.IsSourceDestChecked = nil
	numSpotVMUpdate.SecurityGroupIds = nil
	numSpotVMUpdate.DeletionProtection = nil
	numSpotVMUpdate.UserData = nil
	numSpotVMUpdate.Type = nil
	numSpotVMUpdate.SecurityGroupIds = nil
	numSpotVMUpdate.VmInitiatedShutdownBehavior = nil
	numSpotVMUpdate.NestedVirtualization = nil

	var updateVMResponse *api.UpdateVmResponse
	if updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, spaceID, vmID, *numSpotVMUpdate); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	if err = utils.ParseHTTPError(updateVMResponse.Body, updateVMResponse.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err = StartVM(ctx, provider, vmID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return RetryReadVM(ctx, provider, createOp, vmID)
}

// UpdateVMTags updates VM tags.
func UpdateVMTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	vmID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (numSpotVM *api.Vm, err error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, vmID); err != nil {
		return nil, err
	}
	return RetryReadVM(ctx, provider, updateOp, vmID)
}

// DeleteVM deletes VM.
func DeleteVM(ctx context.Context, provider *client.NumSpotSDK, vmID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, vmID, numspotClient.DeleteVmsWithResponse)
	if err != nil {
		return utils.WrapError(err, "failed to delete resource")
	}

	return nil
}

// RetryReadVM retries read VM.
func RetryReadVM(ctx context.Context, provider *client.NumSpotSDK, op, vmID string) (*api.Vm, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(ctx, vmID, provider.SpaceID, vmPendingStates, vmTargetStates, numspotClient.ReadVmsByIdWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numSpotVM, assert := read.(*api.Vm)
	if !assert {
		return nil, fmt.Errorf("%w %s: %s", errInvalidVMAssertion, vmID, op)
	}
	return numSpotVM, utils.WrapError(err, "failed to wait for resource state")
}

// ReadVM reads VM.
func ReadVM(ctx context.Context, provider *client.NumSpotSDK, vmID string) (*api.Vm, error) {
	var numSpotReadVM *api.ReadVmsByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadVM, err = numspotClient.ReadVmsByIdWithResponse(ctx, provider.SpaceID, vmID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vms by id")
	}
	if err = utils.ParseHTTPError(numSpotReadVM.Body, numSpotReadVM.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vms by id")
	}

	return numSpotReadVM.JSON200, utils.WrapError(err, "failed to read vms by id")
}

// ReadVMsWithParams reads VMs with params.
func ReadVMsWithParams(ctx context.Context, provider *client.NumSpotSDK, params *api.ReadVmsParams) (*[]api.Vm, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadVM, err := numspotClient.ReadVmsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vms")
	}
	if err = utils.ParseHTTPError(numSpotReadVM.Body, numSpotReadVM.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vms")
	}

	if numSpotReadVM.JSON200.Items == nil {
		return nil, errExpectedVMList
	}

	return numSpotReadVM.JSON200.Items, utils.WrapError(err, "failed to read vms")
}

// StopVM handles stop VM.
func StopVM(ctx context.Context, provider *client.NumSpotSDK, vm string) (err error) {
	if vm == "" {
		return nil
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	var vmStatus *api.ReadVmsByIdResponse
	if vmStatus, err = numspotClient.ReadVmsByIdWithResponse(ctx, provider.SpaceID, vm); err != nil {
		return utils.WrapError(err, "failed to get API client")
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
	if _, err = numspotClient.StopVmWithResponse(ctx, provider.SpaceID, vm, api.StopVm{ForceStop: &forceStop}); err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	if _, err = utils.RetryReadUntilStateValid(ctx, vm, provider.SpaceID, []string{stopping}, []string{stopped, terminated},
		numspotClient.ReadVmsByIdWithResponse); err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	return nil
}

// StartVM handles start VM.
func StartVM(ctx context.Context, provider *client.NumSpotSDK, vm string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	// Already running
	if vm == "" {
		return nil
	}

	var vmStatus *api.ReadVmsByIdResponse
	if vmStatus, err = numspotClient.ReadVmsByIdWithResponse(ctx, provider.SpaceID, vm); err != nil {
		return utils.WrapError(err, "failed to get API client")
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
		return utils.WrapError(err, "failed to get API client")
	}

	if _, err = utils.RetryReadUntilStateValid(
		ctx,
		vm,
		provider.SpaceID,
		[]string{pending},
		[]string{running},
		numspotClient.ReadVmsByIdWithResponse,
	); err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	return nil
}
