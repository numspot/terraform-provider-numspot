package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateVolume(ctx context.Context, provider services.IProvider, input numspot.CreateVolumeJSONRequestBody, vmID, deviceName string) (numSpotVolume *numspot.Volume, err error) {
	spaceID := provider.GetSpaceID()

	var retryCreate *numspot.CreateVolumeResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, input,
		provider.GetNumspotClient().CreateVolumeWithResponse); err != nil {
		return nil, err
	}

	volumeID := *retryCreate.JSON201.Id

	if vmID != "" {
		err = link(ctx, provider, volumeID, vmID, deviceName)
		if err != nil {
			return nil, err
		}
	}

	var read any
	if read, err = utils.RetryReadUntilStateValid(ctx, volumeID, spaceID, pendingState{creating, updating}, targetState{available, inUse},
		provider.GetNumspotClient().ReadVolumesByIdWithResponse); err != nil {
		return nil, err
	}

	var assert bool
	if numSpotVolume, assert = read.(*numspot.Volume); !assert {
		return nil, errors.New("invalid create volume assertion")
	}
	return numSpotVolume, err
}

func UpdateVolumeAttributes(
	ctx context.Context,
	provider services.IProvider,
	vmID string,
	volumeID string,
	numSpotVolumeUpdate numspot.UpdateVolumeJSONRequestBody,
) (numSpotVolume *numspot.Volume, err error) {
	// stateVmExists := true
	var assert bool

	if vmID != "" {
		// Stop linked VM before unlinking Volume
		if err = vm.StopVmNoDiag(ctx, provider, vmID); err != nil {
			return nil, err
		}
	}
	if _, err = provider.GetNumspotClient().UpdateVolumeWithResponse(ctx, provider.GetSpaceID(), volumeID, numSpotVolumeUpdate); err != nil {
		return nil, err
	}
	if vmID != "" {
		// Start linked VM after updating Volume
		if err = vm.StartVmNoDiag(ctx, provider, vmID); err != nil {
			return nil, err
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
		return nil, errors.New("invalid update volume attributes assertion")
	}

	return numSpotVolume, nil
}

func unlink(ctx context.Context, provider services.IProvider, volumeID, vmID string) (err error) {
	if err = vm.StopVmNoDiag(ctx, provider, vmID); err != nil {
		return err
	}
	if _, err = provider.GetNumspotClient().UnlinkVolumeWithResponse(
		ctx,
		provider.GetSpaceID(),
		volumeID,
		numspot.UnlinkVolumeJSONRequestBody{},
	); err != nil {
		return err
	}
	if err = vm.StartVmNoDiag(ctx, provider, vmID); err != nil {
		return err
	}
	return nil
}

func link(ctx context.Context, provider services.IProvider, volumeID string, vmID, deviceName string) (err error) {
	spaceID := provider.GetSpaceID()
	linkBody := numspot.LinkVolumeJSONRequestBody{
		DeviceName: deviceName,
		VmId:       vmID,
	}

	if _, err = utils.RetryLinkUntilResourceAvailableWithBody(
		ctx, spaceID, volumeID, linkBody, provider.GetNumspotClient().LinkVolumeWithResponse); err != nil {
		return err
	}

	createStateConf := &retry.StateChangeConf{
		Pending: []string{attaching},
		Target:  []string{attached},
		Refresh: func() (interface{}, string, error) {
			var readRes *numspot.ReadVolumesByIdResponse
			readRes, err = provider.GetNumspotClient().ReadVolumesByIdWithResponse(ctx, spaceID, volumeID)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read volume : %v", err)
			}

			if len(*readRes.JSON200.LinkedVolumes) > 0 {
				linkState := (*readRes.JSON200.LinkedVolumes)[0].State
				return readRes.JSON200, *linkState, nil
			}
			return nil, "", fmt.Errorf("Volume not linked to any VM : %v", err)
		},
		Timeout: utils.TfRequestRetryTimeout,
		Delay:   utils.TfRequestRetryDelay,
	}
	_, err = createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func UpdateVolumeLink(ctx context.Context, provider services.IProvider,
	volumeID, stateVMID, planVMID, planDeviceName string,
) (numSpotVolume *numspot.Volume, err error) {
	var assert bool

	if stateVMID == "" {
		if planVMID != "" {
			// Nothing in the state and VM in the plan
			// We link the volume to the VM in the plan
			if err = link(ctx, provider, volumeID, planVMID, planDeviceName); err != nil {
				return nil, err
			}
		}
	}

	if stateVMID != "" {
		if planVMID == "" {
			// Nothing in the plan and VM in the state
			// We need to unlink the volume to the VM in state
			if err = unlink(ctx, provider, volumeID, stateVMID); err != nil {
				return nil, err
			}
		} else {
			// VM in the state, VM in the plan
			// We need to unlink the volume from the previous VM (in state) and link it to the new VM (in plan) with the device name in the plan
			if err = unlink(ctx, provider, volumeID, stateVMID); err != nil {
				return nil, err
			}
			if err = link(ctx, provider, volumeID, planVMID, planDeviceName); err != nil {
				return nil, err
			}
		}
	}

	retryRead, err := utils.RetryReadUntilStateValid(ctx, volumeID, provider.GetSpaceID(), pendingState{creating, updating}, targetState{available, inUse},
		provider.GetNumspotClient().ReadVolumesByIdWithResponse,
	)
	if err != nil {
		return nil, err
	}

	if numSpotVolume, assert = retryRead.(*numspot.Volume); !assert {
		return nil, errors.New("invalid update link assertion")
	}

	return numSpotVolume, nil
}

func DeleteVolume(ctx context.Context,
	provider services.IProvider,
	volumeID, vmID string,
) (err error) {

	if vmID != "" {
		if err = vm.StopVmNoDiag(ctx, provider, vmID); err != nil {
			return err
		}
		if _, err = provider.GetNumspotClient().UnlinkVolumeWithResponse(
			ctx,
			provider.GetSpaceID(),
			volumeID,
			numspot.UnlinkVolumeJSONRequestBody{},
		); err != nil {
			return err
		}
		if err = vm.StartVmNoDiag(ctx, provider, vmID); err != nil {
			return err
		}
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.GetSpaceID(), volumeID, provider.GetNumspotClient().DeleteVolumeWithResponse)
	if err != nil {
		return err
	}

	return nil
}
