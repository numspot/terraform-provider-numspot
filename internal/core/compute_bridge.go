package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

// CreateComputeBridge creates compute bridge.
func CreateComputeBridge(
	ctx context.Context,
	provider *client.NumSpotSDK,
	createComputeBridge *api.CreateComputeBridgeRequest,
) (*api.CreateComputeBridge201Response, error) {
	numsClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := numsClient.CreateComputeBridgeWithResponse(ctx, provider.SpaceID, *createComputeBridge)
	if err != nil {
		return nil, utils.WrapError(err, "failed to create compute bridge")
	}

	err = utils.ParseHTTPError(res.Body, res.StatusCode())
	if err != nil {
		return nil, utils.WrapError(err, "failed to parse HTTP error response")
	}

	return res.JSON201, nil
}

// DeleteComputeBridge deletes compute bridge.
func DeleteComputeBridge(ctx context.Context, provider *client.NumSpotSDK, computeBridgeID string) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.DeleteComputeBridgeWithResponse(ctx, spaceID, utils.StringToUuid(computeBridgeID))
	if err != nil {
		return utils.WrapError(err, "failed to delete compute bridge")
	}
	err = utils.ParseHTTPError(res.Body, res.StatusCode())
	if err != nil {
		return utils.WrapError(err, "failed to parse HTTP error response")
	}

	return nil
}

// ReadComputeBridges reads compute bridges.
func ReadComputeBridges(ctx context.Context, provider *client.NumSpotSDK) (*[]api.ComputeBridge, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := numspotClient.ListComputeBridgesWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list compute bridges")
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list compute bridges")
	}

	return &read.JSON200.Items, nil
}

// ReadComputeBridge reads compute bridge.
func ReadComputeBridge(ctx context.Context, provider *client.NumSpotSDK, computeBridgeID string) (*api.ComputeBridge, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := numspotClient.ReadComputeBridgeWithResponse(ctx, provider.SpaceID, utils.StringToUuid(computeBridgeID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to read compute bridge")
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read compute bridge")
	}

	return read.JSON200, nil
}
