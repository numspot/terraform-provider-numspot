package core

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	volumePendingStates = []string{creating, updating}
	volumeTargetStates  = []string{available, inUse}
)

func CreateVolume(ctx context.Context, provider *client.NumSpotSDK, numSpotVolumeCreate numspot.CreateVolumeJSONRequestBody, tags []numspot.ResourceTag, vmID, deviceName string) (numSpotVolume *numspot.Volume, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreateVolumeResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotVolumeCreate, numspotClient.CreateVolumeWithResponse); err != nil {
		return nil, err
	}

	volumeID := *retryCreate.JSON201.Id

	if vmID != "" {
		err = linkVolume(ctx, provider, createOp, volumeID, vmID, deviceName)
		if err != nil {
			return nil, err
		}
	}

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, volumeID, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadVolume(ctx, provider, createOp, volumeID)
}

func UpdateVolumeAttributes(ctx context.Context, provider *client.NumSpotSDK, numSpotVolumeUpdate numspot.UpdateVolumeJSONRequestBody, volumeID, stateVM, planVM string) (*numspot.Volume, error) {
	var err error

	// If this volume is attached to a VM, we need to change it from hot to cold volume to update its attributes
	// To make it a cold volume we need to stop the VM it's attached to
	if err = StopVM(ctx, provider, stateVM); err != nil {
		return nil, err
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	var updateVolumeResponse *numspot.UpdateVolumeResponse
	if updateVolumeResponse, err = numspotClient.UpdateVolumeWithResponse(ctx, provider.SpaceID, volumeID, numSpotVolumeUpdate); err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(updateVolumeResponse.Body, updateVolumeResponse.StatusCode()); err != nil {
		return nil, err
	}

	// Starting back up the VM making it a hot volume
	if err = StartVM(ctx, provider, stateVM); err != nil {
		return nil, err
	}

	return RetryReadVolume(ctx, provider, updateOp, volumeID)
}

func UpdateVolumeTags(ctx context.Context, provider *client.NumSpotSDK, volumeID string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.Volume, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, volumeID); err != nil {
		return nil, err
	}
	return RetryReadVolume(ctx, provider, updateOp, volumeID)
}

func UpdateVolumeLink(ctx context.Context, provider *client.NumSpotSDK, volumeID, stateVM, planVM, planDeviceName string) (*numspot.Volume, error) {
	var err error

	switch {
	// Nothing in the state and VM in the plan
	// We link the volume to the VM in the plan
	case stateVM == "" && planVM != "":
		if err = linkVolume(ctx, provider, updateOp, volumeID, planVM, planDeviceName); err != nil {
			return nil, err
		}

	// Nothing in the plan and VM in the state
	// We need to unlink the volume to the VM in state
	case stateVM != "" && planVM == "":
		if err = unlinkVolume(ctx, provider, volumeID, stateVM); err != nil {
			return nil, err
		}

	// VM in the state, VM in the plan
	// We need to unlink the volume from the previous VM (in state) and link it to the new VM (in plan) with the device name in the plan
	case stateVM != "":
		if err = unlinkVolume(ctx, provider, volumeID, stateVM); err != nil {
			return nil, err
		}
		if err = linkVolume(ctx, provider, updateOp, volumeID, planVM, planDeviceName); err != nil {
			return nil, err
		}
	}

	//if stateVM != planVM {
	//	switch {
	//	case stateVM != "":
	//		if err = unlinkVolume(ctx, provider, volumeID, stateVM); err != nil {
	//			return nil, err
	//		}
	//	case planVM != "":
	//		if err = linkVolume(ctx, provider, updateOp, volumeID, planVM, planDeviceName); err != nil {
	//			return nil, err
	//		}
	//	}
	//}

	return RetryReadVolume(ctx, provider, updateOp, volumeID)
}

func DeleteVolume(ctx context.Context, provider *client.NumSpotSDK, volumeID, stateVM string) (err error) {
	if stateVM != "" {
		if err = unlinkVolume(ctx, provider, volumeID, stateVM); err != nil {
			return err // TODO : remove and try to delete volume anyway ?
		}
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, volumeID, numspotClient.DeleteVolumeWithResponse)
	if err != nil {
		return err
	}

	return nil
}

func RetryReadVolume(ctx context.Context, provider *client.NumSpotSDK, op string, volumeID string) (*numspot.Volume, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, volumeID, provider.SpaceID, volumePendingStates, volumeTargetStates, numspotClient.ReadVolumesByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotVolume, assert := read.(*numspot.Volume)
	if !assert {
		return nil, fmt.Errorf("invalid volume assertion %s: %s", volumeID, op)
	}
	return numSpotVolume, err
}

func ReadVolume(ctx context.Context, provider *client.NumSpotSDK, volumeID string) (numSpotVolume *numspot.Volume, err error) {
	var numSpotReadVolume *numspot.ReadVolumesByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadVolume, err = numspotClient.ReadVolumesByIdWithResponse(ctx, provider.SpaceID, volumeID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadVolume.Body, numSpotReadVolume.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadVolume.JSON200, err
}

func unlinkVolume(ctx context.Context, provider *client.NumSpotSDK, volumeID, stateVM string) (err error) {
	volume, err := ReadVolume(ctx, provider, volumeID)
	if err != nil {
		return err
	}

	if *volume.State == inUse {
		if err = StopVM(ctx, provider, stateVM); err != nil {
			return err
		}

		var numSpotClient *numspot.ClientWithResponses
		numSpotClient, err = provider.GetClient(ctx)
		if err != nil {
			return err
		}
		var unlinkVolumeResponse *numspot.UnlinkVolumeResponse
		if unlinkVolumeResponse, err = numSpotClient.UnlinkVolumeWithResponse(ctx, provider.SpaceID, volumeID, numspot.UnlinkVolumeJSONRequestBody{}); err != nil {
			return err
		}
		if err = utils.ParseHTTPError(unlinkVolumeResponse.Body, unlinkVolumeResponse.StatusCode()); err != nil {
			return err
		}

		if err = StartVM(ctx, provider, stateVM); err != nil {
			return err
		}
	}

	return nil
}

func linkVolume(ctx context.Context, provider *client.NumSpotSDK, op, volumeID, vmID, deviceName string) (err error) {
	spaceID := provider.SpaceID
	linkBody := numspot.LinkVolumeJSONRequestBody{
		DeviceName: deviceName,
		VmId:       vmID,
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	if _, err = utils.RetryLinkUntilResourceAvailableWithBody(ctx, spaceID, volumeID, linkBody, numspotClient.LinkVolumeWithResponse); err != nil {
		return err
	}

	createStateConf := &retry.StateChangeConf{
		Pending: []string{attaching},
		Target:  []string{attached},
		Timeout: utils.TfRequestRetryTimeout,
		Delay:   utils.ParseRetryBackoff(),
		Refresh: func() (interface{}, string, error) {
			var volume *numspot.Volume
			if volume, err = RetryReadVolume(ctx, provider, op, volumeID); err != nil {
				return nil, "", err
			}

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
