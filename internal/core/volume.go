package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	volumePendingStates = []string{creating, updating}
	volumeTargetStates  = []string{available, inUse}

	// errVolumeDownsizeNotAllowed is returned when an update attempts to shrink
	// a volume without the replace-on-downsize option.
	errVolumeDownsizeNotAllowed = errors.New("trying to update volume size")
	// errInvalidVolumeAssertion is returned when a volume read result cannot be
	// asserted to the expected type.
	errInvalidVolumeAssertion = errors.New("invalid volume assertion")
	// errExpectedVolumeList is returned when a volume list response is
	// unexpectedly nil.
	errExpectedVolumeList = errors.New("HTTP call failed : expected a list of volumes but got nil")
	// errVolumeNotLinkedToVM is returned when a volume has no VM attachment.
	errVolumeNotLinkedToVM = errors.New("volume not linked to any vm")
)

// CreateVolume creates volume.
func CreateVolume(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotVolumeCreate *api.CreateVolumeJSONRequestBody,
	vmID, deviceName string,
	tags []api.ResourceTag,
) (numspotVolume *api.Volume, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateVolumeResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*numspotVolumeCreate,
		numspotClient.CreateVolumeWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	volumeID := *retryCreate.JSON201.Id

	if vmID != "" {
		if err := linkVolume(ctx, provider, createOp, volumeID, vmID, deviceName); err != nil {
			return nil, err
		}
	}

	if len(tags) > 0 {
		if err := createTags(ctx, provider, volumeID, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadVolume(ctx, provider, createOp, volumeID)
}

// UpdateVolumeAttributes updates volume attributes.
func UpdateVolumeAttributes(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotVolumeUpdate *api.UpdateVolumeJSONRequestBody,
	volumeID, stateVM string,
	stateSize int64,
) (*api.Volume, error) {
	// Downsizing a volume in place is not supported by the API. The
	// requires_replace_if plan modifier forces a replacement when downsizing with
	// replace_volume_on_downsize set; reaching here with a smaller size means the
	// flag is off, so reject it.
	if numspotVolumeUpdate.Size != nil && int64(*numspotVolumeUpdate.Size) < stateSize {
		return nil, fmt.Errorf("%w from %v to %v. It is not possible to downsize a volume in an update. "+
			"To force the replace of volume, set attribute 'replace_volume_on_downsize' to true. Note : all data on volume will be lost",
			errVolumeDownsizeNotAllowed, stateSize, *numspotVolumeUpdate.Size)
	}

	// If this volume is attached to a VM, we need to change it from hot to cold volume to update its attributes
	// To make it a cold volume we need to stop the VM it's attached to
	if err := StopVM(ctx, provider, stateVM); err != nil {
		return nil, err
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	var updateVolumeResponse *api.UpdateVolumeResponse
	if updateVolumeResponse, err = numspotClient.UpdateVolumeWithResponse(ctx, provider.SpaceID, volumeID, *numspotVolumeUpdate); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	if err = utils.ParseHTTPError(updateVolumeResponse.Body, updateVolumeResponse.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	// Starting back up the VM making it a hot volume
	if err = StartVM(ctx, provider, stateVM); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return RetryReadVolume(ctx, provider, updateOp, volumeID)
}

// UpdateVolumeTags updates volume tags.
func UpdateVolumeTags(
	ctx context.Context,
	provider *client.NumspotSDK,
	volumeID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.Volume, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, volumeID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	return RetryReadVolume(ctx, provider, updateOp, volumeID)
}

// UpdateVolumeLink updates volume link.
func UpdateVolumeLink(ctx context.Context, provider *client.NumspotSDK, volumeID, stateVM, planVM, planDeviceName string) (*api.Volume, error) {
	switch {
	// Nothing in the state and VM in the plan
	// We link the volume to the VM in the plan
	case stateVM == "" && planVM != "":
		if err := linkVolume(ctx, provider, updateOp, volumeID, planVM, planDeviceName); err != nil {
			return nil, err
		}

	// Nothing in the plan and VM in the state
	// We need to unlink the volume to the VM in state
	case stateVM != "" && planVM == "":
		if err := unlinkVolume(ctx, provider, volumeID, stateVM); err != nil {
			return nil, err
		}

	// VM in the state, VM in the plan
	// We need to unlink the volume from the previous VM (in state) and link it to the new VM (in plan) with the device name in the plan
	case stateVM != "":
		if err := unlinkVolume(ctx, provider, volumeID, stateVM); err != nil {
			return nil, err
		}
		if err := linkVolume(ctx, provider, updateOp, volumeID, planVM, planDeviceName); err != nil {
			return nil, err
		}
	}

	return RetryReadVolume(ctx, provider, updateOp, volumeID)
}

// DeleteVolume deletes volume.
func DeleteVolume(ctx context.Context, provider *client.NumspotSDK, volumeID, stateVM string) (err error) {
	if stateVM != "" {
		if err := unlinkVolume(ctx, provider, volumeID, stateVM); err != nil {
			return err // TODO : remove and try to delete volume anyway ?
		}
	}
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, volumeID, numspotClient.DeleteVolumeWithResponse),
		"failed to delete resource",
	)
}

// RetryReadVolume retries read volume.
func RetryReadVolume(ctx context.Context, provider *client.NumspotSDK, op, volumeID string) (*api.Volume, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		volumeID,
		provider.SpaceID,
		volumePendingStates,
		volumeTargetStates,
		numspotClient.ReadVolumesByIdWithResponse,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numspotVolume, assert := read.(*api.Volume)
	if !assert {
		return nil, fmt.Errorf("%w %s: %s", errInvalidVolumeAssertion, volumeID, op)
	}
	return numspotVolume, utils.WrapError(err, "failed to wait for resource state")
}

// ReadVolume reads volume.
func ReadVolume(ctx context.Context, provider *client.NumspotSDK, volumeID string) (numspotVolume *api.Volume, err error) {
	var numspotReadVolume *api.ReadVolumesByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadVolume, err = numspotClient.ReadVolumesByIdWithResponse(ctx, provider.SpaceID, volumeID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read volumes by id")
	}
	if err = utils.ParseHTTPError(numspotReadVolume.Body, numspotReadVolume.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read volumes by id")
	}

	return numspotReadVolume.JSON200, nil
}

// ReadVolumesWithParams reads volumes with params.
func ReadVolumesWithParams(ctx context.Context, provider *client.NumspotSDK, params *api.ReadVolumesParams) (numspotVolume *[]api.Volume, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadVolume, err := numspotClient.ReadVolumesWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read volumes")
	}
	if err = utils.ParseHTTPError(numspotReadVolume.Body, numspotReadVolume.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read volumes")
	}
	if numspotReadVolume.JSON200.Items == nil {
		return nil, errExpectedVolumeList
	}

	return numspotReadVolume.JSON200.Items, nil
}

func unlinkVolume(ctx context.Context, provider *client.NumspotSDK, volumeID, stateVM string) (err error) {
	volume, err := ReadVolume(ctx, provider, volumeID)
	if err != nil {
		return err
	}

	if *volume.State == inUse {
		if err := StopVM(ctx, provider, stateVM); err != nil {
			return err
		}

		numspotClient, err := provider.GetClient(ctx)
		if err != nil {
			return utils.WrapError(err, "failed to get API client")
		}
		unlinkVolumeResponse, err := numspotClient.UnlinkVolumeWithResponse(
			ctx,
			provider.SpaceID,
			volumeID,
			api.UnlinkVolumeJSONRequestBody{},
		)
		if err != nil {
			return utils.WrapError(err, "failed to get API client")
		}
		if err = utils.ParseHTTPError(unlinkVolumeResponse.Body, unlinkVolumeResponse.StatusCode()); err != nil {
			return utils.WrapError(err, "failed to get API client")
		}

		if err := StartVM(ctx, provider, stateVM); err != nil {
			return utils.WrapError(err, "failed to get API client")
		}
	}

	return nil
}

func linkVolume(ctx context.Context, provider *client.NumspotSDK, op, volumeID, vmID, deviceName string) (err error) {
	spaceID := provider.SpaceID
	linkBody := api.LinkVolumeJSONRequestBody{
		DeviceName: deviceName,
		VmId:       vmID,
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	if _, err = utils.RetryUntilResourceAvailableWithBody(ctx, spaceID, volumeID, linkBody, numspotClient.LinkVolumeWithResponse); err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	createStateConf := &retry.StateChangeConf{
		Pending: []string{attaching},
		Target:  []string{attached},
		Timeout: utils.TfRequestRetryTimeout,
		Delay:   utils.ParseRetryBackoff(),
		Refresh: func() (interface{}, string, error) {
			var volume *api.Volume
			if volume, err = RetryReadVolume(ctx, provider, op, volumeID); err != nil {
				return nil, "", utils.WrapError(err, "failed to get API client")
			}

			if len(*volume.LinkedVolumes) > 0 {
				linkState := (*volume.LinkedVolumes)[0].State
				return volume, *linkState, nil
			}
			return nil, "", fmt.Errorf("%w: %v", errVolumeNotLinkedToVM, err)
		},
	}

	if _, err = createStateConf.WaitForStateContext(ctx); err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	return nil
}
