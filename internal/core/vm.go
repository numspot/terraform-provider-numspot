package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
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
	provider *client.NumspotSDK,
	numspotVMCreate *api.CreateVmsJSONRequestBody,
	tags []api.ResourceTag,
) (numspotVM *api.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	// BootOnCreation is not a Terraform attribute; the resource always boots the
	// VM on creation, so set it here rather than in the (generated) request body.
	if numspotVMCreate.BootOnCreation == nil {
		numspotVMCreate.BootOnCreation = utils.PointerOf(true)
	}

	var retryCreate *api.CreateVmsResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, *numspotVMCreate,
		numspotClient.CreateVmsWithResponse); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if retryCreate.JSON201 == nil || retryCreate.JSON201.Id == nil {
		return nil, fmt.Errorf("%w (HTTP %d)", errMissingCreateID, retryCreate.StatusCode())
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
	provider *client.NumspotSDK,
	numspotVMUpdate *api.UpdateVmJSONRequestBody,
	vmID string,
) (numspotVM *api.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err = StopVM(ctx, provider, vmID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numspotVMUpdate.KeypairName = nil

	var updateVMResponse *api.UpdateVmResponse
	if updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, spaceID, vmID, *numspotVMUpdate); err != nil {
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
// so a SG-only change must not force the
// stop/update/start cycle of UpdateVMAttributes (CLOS-3261).
func UpdateVMSecurityGroups(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotVMUpdate *api.UpdateVmJSONRequestBody,
	vmID string,
) (numspotVM *api.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numspotVMUpdate.KeypairName = nil
	numspotVMUpdate.BlockDeviceMappings = nil
	numspotVMUpdate.DeletionProtection = nil
	numspotVMUpdate.NestedVirtualization = nil
	numspotVMUpdate.Type = nil
	numspotVMUpdate.UserData = nil
	numspotVMUpdate.VmInitiatedShutdownBehavior = nil
	numspotVMUpdate.IsSourceDestChecked = nil

	var updateVMResponse *api.UpdateVmResponse
	if updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, spaceID, vmID, *numspotVMUpdate); err != nil {
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
	provider *client.NumspotSDK,
	numspotVMUpdate *api.UpdateVmJSONRequestBody,
	vmID string,
) (numspotVM *api.Vm, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err = StopVM(ctx, provider, vmID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numspotVMUpdate.BlockDeviceMappings = nil
	numspotVMUpdate.IsSourceDestChecked = nil
	numspotVMUpdate.SecurityGroupIds = nil
	numspotVMUpdate.DeletionProtection = nil
	numspotVMUpdate.UserData = nil
	numspotVMUpdate.Type = nil
	numspotVMUpdate.SecurityGroupIds = nil
	numspotVMUpdate.VmInitiatedShutdownBehavior = nil
	numspotVMUpdate.NestedVirtualization = nil

	var updateVMResponse *api.UpdateVmResponse
	if updateVMResponse, err = numspotClient.UpdateVmWithResponse(ctx, spaceID, vmID, *numspotVMUpdate); err != nil {
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
	provider *client.NumspotSDK,
	vmID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (numspotVM *api.Vm, err error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, vmID); err != nil {
		return nil, err
	}
	return RetryReadVM(ctx, provider, updateOp, vmID)
}

// DeleteVM deletes VM.
func DeleteVM(ctx context.Context, provider *client.NumspotSDK, vmID string) (err error) {
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
func RetryReadVM(ctx context.Context, provider *client.NumspotSDK, op, vmID string) (*api.Vm, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(ctx, vmID, provider.SpaceID, vmPendingStates, vmTargetStates, numspotClient.ReadVmsByIdWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numspotVM, assert := read.(*api.Vm)
	if !assert {
		return nil, fmt.Errorf("%w %s: %s", errInvalidVMAssertion, vmID, op)
	}
	return numspotVM, utils.WrapError(err, "failed to wait for resource state")
}

// ReadVM reads VM.
func ReadVM(ctx context.Context, provider *client.NumspotSDK, vmID string) (*api.Vm, error) {
	var numspotReadVM *api.ReadVmsByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadVM, err = numspotClient.ReadVmsByIdWithResponse(ctx, provider.SpaceID, vmID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vms by id")
	}
	if err = utils.ParseHTTPError(numspotReadVM.Body, numspotReadVM.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vms by id")
	}

	return numspotReadVM.JSON200, utils.WrapError(err, "failed to read vms by id")
}

// ReadVMsWithParams reads VMs with params.
func ReadVMsWithParams(ctx context.Context, provider *client.NumspotSDK, params *api.ReadVmsParams) (*[]api.Vm, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadVM, err := numspotClient.ReadVmsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vms")
	}
	if err = utils.ParseHTTPError(numspotReadVM.Body, numspotReadVM.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vms")
	}

	if numspotReadVM.JSON200.Items == nil {
		return nil, errExpectedVMList
	}

	return numspotReadVM.JSON200.Items, utils.WrapError(err, "failed to read vms")
}

// StopVM handles stop VM.
func StopVM(ctx context.Context, provider *client.NumspotSDK, vm string) (err error) {
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
func StartVM(ctx context.Context, provider *client.NumspotSDK, vm string) (err error) {
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
