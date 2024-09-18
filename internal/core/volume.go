package core

import (
	"context"
	"errors"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

//////////////////////
// Create Volume
//
// Retry Create
// Retry Read
// If VM >
// 			Link VM
//			Retry Read
//////////////////////
// Read Volume
//
//////////////////////
// Update Volume
//
//////////////////////
// Delete Volume
//
//////////////////////

func CreateVolume(ctx context.Context, provider services.IProvider, input numspot.CreateVolumeJSONRequestBody, vmID, deviceName string) (numSpotVolume *numspot.Volume, err error) {
	var assert bool
	spaceID := provider.GetSpaceID()

	// Retry Create volume
	var retryCreate *numspot.CreateVolumeResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID,
		input,
		provider.GetNumspotClient().CreateVolumeWithResponse); err != nil {
		return nil, err
	}
	volumeID := *retryCreate.JSON201.Id

	// Retry Read - Wait for volume to be ready
	var retryRead any
	if retryRead, err = utils.RetryReadUntilStateValid(ctx, volumeID, spaceID,
		pendingState{creating},
		targetState{available},
		provider.GetNumspotClient().ReadVolumesByIdWithResponse); err != nil {
		return nil, err
	}
	if numSpotVolume, assert = retryRead.(*numspot.Volume); !assert {
		return nil, errors.New("invalid read assertion")
	}

	// Link current volume to a VM is specified
	if vmID != "" {
		linkBody := numspot.LinkVolumeJSONRequestBody{
			DeviceName: deviceName,
			VmId:       vmID,
		}
		if _, err = provider.GetNumspotClient().LinkVolumeWithResponse(ctx, spaceID, volumeID, linkBody); err != nil {
			return nil, err
		}

		// Retry Read - Wait for volume to be linked to VM
		var retryReadLink any
		if retryReadLink, err = utils.RetryReadUntilStateValid(ctx, volumeID, spaceID,
			pendingState{available},
			targetState{inUse},
			provider.GetNumspotClient().ReadVolumesByIdWithResponse); err != nil {
			return nil, err
		}
		if numSpotVolume, assert = retryReadLink.(*numspot.Volume); !assert {
			return nil, errors.New("invalid read link assertion")
		}
	}

	return numSpotVolume, err
}

func UpdateVolumeAttributes(
	ctx context.Context,
	provider services.IProvider,
	volumeID string,
	numSpotVolumeUpdate numspot.UpdateVolumeJSONRequestBody,
	vmID, deviceName string,
) (numSpotVolume *numspot.Volume, err error) {
	// stateVmExists := true
	var assert bool
	if vmID != "" {
		// Stop linked VM before unlinking Volume
		if err, _ = stopVm(ctx, provider, vmID); err != nil {
			return nil, err
		}
	}

	if _, err = provider.GetNumspotClient().UpdateVolumeWithResponse(ctx, provider.GetSpaceID(), volumeID, numSpotVolumeUpdate); err != nil {
		return nil, err
	}

	if vmID != "" {
		// Start linked VM after updating Volume
		diags := vm.StartVm(ctx, provider, vmID)
		if diags.HasError() {
			return nil, errors.New("failed to start vm")
		}
	}

	// Retries read on resource until state is OK
	var retryRead any
	if retryRead, err = utils.RetryReadUntilStateValid(
		ctx,
		volumeID,
		provider.GetSpaceID(),
		pendingState{creating, updating},
		targetState{available, inUse},
		provider.GetNumspotClient().ReadVolumesByIdWithResponse,
	); err != nil {
		return nil, err
	}

	if numSpotVolume, assert = retryRead.(*numspot.Volume); !assert {
		return nil, errors.New("invalid read update assertion")
	}

	return numSpotVolume, nil
}

func UpdateVolumeLink(
	ctx context.Context,
	provider services.IProvider,
	volumeID, vmOldID, vmNewID, deviceName string,
) (numSpotVolume *numspot.Volume, err error) {
	var assert bool
	// stateVmExists := true
	if vmOldID != "" {
		// Stop linked VM before unlinking Volume
		if err, _ = stopVm(ctx, provider, vmOldID); err != nil {
			return nil, err
		}
	}

	_, err = provider.GetNumspotClient().UnlinkVolumeWithResponse(
		ctx,
		provider.GetSpaceID(),
		volumeID,
		numspot.UnlinkVolumeJSONRequestBody{},
	)

	if vmOldID != "" {
		// Start linked VM after updating Volume
		if err, _ = startVm(ctx, provider, vmOldID); err != nil {
			return nil, err
		}
	}

	_, err = provider.GetNumspotClient().LinkVolumeWithResponse(
		ctx,
		provider.GetSpaceID(),
		volumeID,
		numspot.LinkVolumeJSONRequestBody{
			DeviceName: vmNewID,
			VmId:       deviceName,
		},
	)

	// Retries read on resource until state is OK
	retryRead, err := utils.RetryReadUntilStateValid(
		ctx,
		volumeID,
		provider.GetSpaceID(),
		pendingState{creating, updating},
		targetState{available, inUse},
		provider.GetNumspotClient().ReadVolumesByIdWithResponse,
	)
	if err != nil {
		//response.Diagnostics.AddError("Failed to update volume", fmt.Sprintf("Error waiting for volume (%s) to be created: %s", state.Id.ValueString(), err))
		return nil, err
	}

	if numSpotVolume, assert = retryRead.(*numspot.Volume); !assert {
		return nil, errors.New("invalid update read assertion")
	}

	return numSpotVolume, nil
}

func stopVm(ctx context.Context, provider services.IProvider, vmId string) (error, bool) {
	diags := vm.StopVm(ctx, provider, vmId)
	if diags.HasError() {
		if !vm.VmIsDeleted(vmId) {
			return errors.New("unable to stop VM"), true
		} else { // Else, the VM got removed, which is OK
			return errors.New("unable to stop VM"), false
		}
	}

	return nil, true
}

func startVm(ctx context.Context, provider services.IProvider, vmId string) (error, bool) {
	diags := vm.StartVm(ctx, provider, vmId)
	if diags.HasError() {
		if !vm.VmIsDeleted(vmId) {
			return errors.New("unable to start VM"), true
		} else { // Else, the VM got removed, which is OK
			return errors.New("unable to start VM"), false
		}
	}

	return nil, true
}

//func linkVmConfigured(data VolumeModel) bool {
//	return !utils.IsTfValueNull(data.LinkVM) && !utils.IsTfValueNull(data.LinkVM.VmID)
//}
