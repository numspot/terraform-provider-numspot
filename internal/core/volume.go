package core

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateVolume(ctx context.Context, provider services.IProvider, numSpotVolumeCreate numspot.CreateVolumeJSONRequestBody, tags []numspot.ResourceTag, vmID, deviceName string) (numSpotVolume *numspot.Volume, err error) {
	spaceID := provider.GetSpaceID()

	var retryCreate *numspot.CreateVolumeResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotVolumeCreate,
		provider.GetNumspotClient().CreateVolumeWithResponse); err != nil {
		return nil, err
	}

	volumeID := *retryCreate.JSON201.Id

	if vmID != "" {
		err = linkVolume(ctx, provider, pendingState{creating, updating}, targetState{available, inUse}, createOp, volumeID, vmID, deviceName)
		if err != nil {
			return nil, err
		}
	}

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider.GetNumspotClient(), spaceID, volumeID, tags); err != nil {
			return nil, err
		}
	}

	return ReadVolume(ctx, provider, pendingState{creating, updating}, targetState{available, inUse}, createOp, volumeID)
}

func UpdateVolumeAttributes(ctx context.Context, provider services.IProvider, numSpotVolumeUpdate numspot.UpdateVolumeJSONRequestBody, volumeID, stateVM, planVM string) (*numspot.Volume, error) {
	var err error
	pendingStates := pendingState{creating, updating}
	targetStates := targetState{available, inUse}

	// If this volume is attached to a VM, we need to change it from hot to cold volume to update its attributes
	// To make it a cold volume we need to stop the VM it's attached to
	if err = vm.StopVmNoDiag(ctx, provider, stateVM); err != nil {
		return nil, err
	}
	if err = vm.StopVmNoDiag(ctx, provider, planVM); err != nil {
		return nil, err
	}

	if _, err = provider.GetNumspotClient().UpdateVolumeWithResponse(ctx, provider.GetSpaceID(), volumeID, numSpotVolumeUpdate); err != nil {
		return nil, err
	}

	// Starting back up the VM making it a hot volume
	if err = vm.StartVmNoDiag(ctx, provider, stateVM); err != nil {
		return nil, err
	}
	if err = vm.StartVmNoDiag(ctx, provider, planVM); err != nil {
		return nil, err
	}

	return ReadVolume(ctx, provider, pendingStates, targetStates, updateOp, volumeID)
}

func UpdateVolumeTags(ctx context.Context, provider services.IProvider, volumeID string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.Volume, error) {
	pendingStates := pendingState{creating, updating}
	targetStates := targetState{available, inUse}

	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, volumeID); err != nil {
		return nil, err
	}
	return ReadVolume(ctx, provider, pendingStates, targetStates, updateOp, volumeID)
}

func UpdateVolumeLink(ctx context.Context, provider services.IProvider, volumeID, stateVM, planVM, planDeviceName string) (*numspot.Volume, error) {
	var err error
	pendingStates := pendingState{creating, updating}
	targetStates := targetState{available, inUse}

	if stateVM == "" && planVM != "" {
		// Nothing in the state and VM in the plan
		// We link the volume to the VM in the plan
		if err = linkVolume(ctx, provider, pendingStates, targetStates, updateOp, volumeID, planVM, planDeviceName); err != nil {
			return nil, err
		}
	}

	if stateVM != "" {
		if planVM == "" {
			// Nothing in the plan and VM in the state
			// We need to unlink the volume to the VM in state
			if err = unlinkVolume(ctx, provider, volumeID, stateVM, planVM); err != nil {
				return nil, err
			}
		} else {
			// VM in the state, VM in the plan
			// We need to unlink the volume from the previous VM (in state) and link it to the new VM (in plan) with the device name in the plan
			if err = unlinkVolume(ctx, provider, volumeID, stateVM, planVM); err != nil {
				return nil, err
			}
			if err = linkVolume(ctx, provider, pendingStates, targetStates, updateOp, volumeID, planVM, planDeviceName); err != nil {
				return nil, err
			}
		}
	}

	return ReadVolume(ctx, provider, pendingStates, targetStates, updateOp, volumeID)
}

func DeleteVolume(ctx context.Context, provider services.IProvider, volumeID, stateVM, planVM string) (err error) {
	if err = unlinkVolume(ctx, provider, volumeID, stateVM, planVM); err != nil {
		return err
	}
	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.GetSpaceID(), volumeID, provider.GetNumspotClient().DeleteVolumeWithResponse)
	if err != nil {
		return err
	}

	return nil
}

func ReadVolume(ctx context.Context, provider services.IProvider, pendingStates pendingState, targetStates targetState, op string, volumeID string) (*numspot.Volume, error) {
	read, err := utils.RetryReadUntilStateValid(ctx, volumeID, provider.GetSpaceID(), pendingStates, targetStates, provider.GetNumspotClient().ReadVolumesByIdWithResponse)
	if err != nil {
		return nil, err
	}
	numSpotVolume, assert := read.(*numspot.Volume)
	if !assert {
		return nil, fmt.Errorf("invalid volume assertion %s: %s", volumeID, op)
	}
	return numSpotVolume, err
}

func unlinkVolume(ctx context.Context, provider services.IProvider, volumeID, stateVM, planVM string) (err error) {
	if err = vm.StopVmNoDiag(ctx, provider, stateVM); err != nil {
		return err
	}
	if err = vm.StopVmNoDiag(ctx, provider, planVM); err != nil {
		return err
	}
	if _, err = provider.GetNumspotClient().UnlinkVolumeWithResponse(ctx, provider.GetSpaceID(), volumeID, numspot.UnlinkVolumeJSONRequestBody{}); err != nil {
		return err
	}
	if err = vm.StartVmNoDiag(ctx, provider, stateVM); err != nil {
		return err
	}
	if err = vm.StartVmNoDiag(ctx, provider, planVM); err != nil {
		return err
	}
	return nil
}

func linkVolume(ctx context.Context, provider services.IProvider, pendingStates pendingState, targetStates targetState, op, volumeID, vmID, deviceName string) (err error) {
	spaceID := provider.GetSpaceID()
	linkBody := numspot.LinkVolumeJSONRequestBody{
		DeviceName: deviceName,
		VmId:       vmID,
	}

	if _, err = utils.RetryLinkUntilResourceAvailableWithBody(ctx, spaceID, volumeID, linkBody, provider.GetNumspotClient().LinkVolumeWithResponse); err != nil {
		return err
	}

	createStateConf := &retry.StateChangeConf{
		Pending: []string{attaching},
		Target:  []string{attached},
		Timeout: utils.TfRequestRetryTimeout,
		Delay:   utils.TfRequestRetryDelay,
		Refresh: func() (interface{}, string, error) {
			var volume *numspot.Volume
			volume, err = ReadVolume(ctx, provider, pendingStates, targetStates, op, volumeID)
			if len(*volume.LinkedVolumes) > 0 {
				linkState := (*volume.LinkedVolumes)[0].State
				return volume, *linkState, nil
			}
			return nil, "", fmt.Errorf("volume not linked to any VM : %v", err)
		},
	}

	if _, err = createStateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}
