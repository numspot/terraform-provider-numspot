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
	flexibleGPUPendingStates = []string{"attaching", "detaching"}
	flexibleGPUTargetStates  = []string{"allocated", "attached"}

	// errExpectedFlexibleGPUList is returned when a flexible GPU list response is
	// unexpectedly nil.
	errExpectedFlexibleGPUList = errors.New("HTTP call failed: expected a list of flexible GPUs but got nil")
	// errInvalidFlexibleGPUAssertion is returned when a flexible GPU read result
	// cannot be asserted to the expected type.
	errInvalidFlexibleGPUAssertion = errors.New("invalid flexible GPU assertion for")
)

// CreateFlexibleGPU creates flexible GPU.
func CreateFlexibleGPU(
	ctx context.Context,
	provider *client.NumSpotSDK,
	body *api.CreateFlexibleGpuJSONRequestBody,
	vmID string,
) (*api.FlexibleGpu, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := utils.RetryCreateUntilResourceAvailableWithBody(ctx, provider.SpaceID, *body, numspotClient.CreateFlexibleGpuWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to create resource")
	}

	gpuID := *res.JSON201.Id

	if vmID != "" {
		if err = LinkFlexibleGPUToVM(ctx, provider, gpuID, vmID); err != nil {
			return nil, utils.WrapError(err, "failed to create resource")
		}
	}

	return retryReadFlexibleGPU(ctx, provider, gpuID)
}

// LinkFlexibleGPUToVM handles link flexible GPU to VM.
func LinkFlexibleGPUToVM(ctx context.Context, provider *client.NumSpotSDK, gpuID, vmID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.LinkFlexibleGpuWithResponse(ctx, provider.SpaceID, gpuID, api.LinkFlexibleGpuJSONRequestBody{VmId: vmID})
	if err != nil {
		return utils.WrapError(err, "failed to link flexible gpu")
	}
	return utils.WrapError(utils.ParseHTTPError(res.Body, res.StatusCode()), "failed to parse HTTP error response")
}

// UnlinkFlexibleGPUFromVM handles unlink flexible GPU from VM.
func UnlinkFlexibleGPUFromVM(ctx context.Context, provider *client.NumSpotSDK, gpuID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.UnlinkFlexibleGpuWithResponse(ctx, provider.SpaceID, gpuID)
	if err != nil {
		return utils.WrapError(err, "failed to unlink flexible gpu")
	}
	return utils.WrapError(utils.ParseHTTPError(res.Body, res.StatusCode()), "failed to parse HTTP error response")
}

// UpdateFlexibleGPU updates flexible GPU.
func UpdateFlexibleGPU(
	ctx context.Context,
	provider *client.NumSpotSDK,
	gpuID, stateVMID, planVMID string,
	deleteOnVMDeletionChanged bool,
	updateBody api.UpdateFlexibleGpuJSONRequestBody,
) (*api.FlexibleGpu, error) {
	if stateVMID != planVMID {
		if stateVMID != "" {
			if err := unlinkFlexibleGPUWithToleratedError(ctx, provider, gpuID); err != nil {
				return nil, utils.WrapError(err, "failed to unlink flexible gpu")
			}
		}
		if planVMID != "" {
			if err := LinkFlexibleGPUToVM(ctx, provider, gpuID, planVMID); err != nil {
				return nil, utils.WrapError(err, "failed to unlink flexible gpu")
			}
		}
	}

	if deleteOnVMDeletionChanged {
		if err := updateFlexibleGPUAttributes(ctx, provider, gpuID, updateBody); err != nil {
			return nil, utils.WrapError(err, "failed to unlink flexible gpu")
		}
	}

	return ReadFlexibleGPU(ctx, provider, gpuID)
}

func updateFlexibleGPUAttributes(ctx context.Context, provider *client.NumSpotSDK, gpuID string, body api.UpdateFlexibleGpuJSONRequestBody) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.UpdateFlexibleGpuWithResponse(ctx, provider.SpaceID, gpuID, body)
	if err != nil {
		return utils.WrapError(err, "failed to update flexible gpu")
	}
	return utils.WrapError(utils.ParseHTTPError(res.Body, res.StatusCode()), "failed to parse HTTP error response")
}

// DeleteFlexibleGPU deletes flexible GPU.
func DeleteFlexibleGPU(ctx context.Context, provider *client.NumSpotSDK, gpuID, vmID string) error {
	if vmID != "" {
		if err := unlinkFlexibleGPUWithToleratedError(ctx, provider, gpuID); err != nil {
			return utils.WrapError(err, "failed to update flexible gpu")
		}
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, gpuID, numspotClient.DeleteFlexibleGpuWithResponse),
		"failed to delete resource",
	)
}

// ReadFlexibleGPU reads flexible GPU.
func ReadFlexibleGPU(ctx context.Context, provider *client.NumSpotSDK, gpuID string) (*api.FlexibleGpu, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.ReadFlexibleGpusByIdWithResponse(ctx, provider.SpaceID, gpuID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read flexible gpus by id")
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read flexible gpus by id")
	}
	return res.JSON200, nil
}

// ReadFlexibleGpusWithParams reads flexible gpus with params.
func ReadFlexibleGpusWithParams(ctx context.Context, provider *client.NumSpotSDK, params *api.ReadFlexibleGpusParams) (*[]api.FlexibleGpu, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.ReadFlexibleGpusWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read flexible gpus")
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read flexible gpus")
	}
	if res.JSON200.Items == nil {
		return nil, errExpectedFlexibleGPUList
	}
	return res.JSON200.Items, nil
}

func retryReadFlexibleGPU(ctx context.Context, provider *client.NumSpotSDK, gpuID string) (*api.FlexibleGpu, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := utils.RetryReadUntilStateValid(
		ctx,
		gpuID,
		provider.SpaceID,
		flexibleGPUPendingStates,
		flexibleGPUTargetStates,
		numspotClient.ReadFlexibleGpusByIdWithResponse,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	gpu, ok := read.(*api.FlexibleGpu)
	if !ok {
		return nil, fmt.Errorf("%w %s", errInvalidFlexibleGPUAssertion, gpuID)
	}
	return gpu, nil
}

func unlinkFlexibleGPUWithToleratedError(ctx context.Context, provider *client.NumSpotSDK, gpuID string) error {
	if err := UnlinkFlexibleGPUFromVM(ctx, provider, gpuID); err == nil {
		return nil
	} else if _, waitErr := retryReadFlexibleGPUForState(ctx, provider, gpuID, []string{"detaching"}, []string{"allocated"}); waitErr != nil {
		return fmt.Errorf("failed to unlink flexible GPU %s: %w; wait for allocated state also failed: %v", gpuID, err, waitErr)
	}
	return nil
}

func retryReadFlexibleGPUForState(
	ctx context.Context,
	provider *client.NumSpotSDK,
	gpuID string,
	pending, target []string,
) (*api.FlexibleGpu, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := utils.RetryReadUntilStateValid(ctx, gpuID, provider.SpaceID, pending, target, numspotClient.ReadFlexibleGpusByIdWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	gpu, ok := read.(*api.FlexibleGpu)
	if !ok {
		return nil, fmt.Errorf("%w %s", errInvalidFlexibleGPUAssertion, gpuID)
	}
	return gpu, nil
}
