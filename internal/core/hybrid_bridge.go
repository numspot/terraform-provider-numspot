package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

// CreateHybridBridge creates hybrid bridge.
func CreateHybridBridge(
	ctx context.Context,
	provider *client.NumSpotSDK,
	createHybridBridgesBridge *api.CreateHybridBridgeRequest,
) (*api.CreateHybridBridge201Response, error) {
	spaceID := provider.SpaceID

	numsClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := numsClient.CreateHybridBridgeWithResponse(ctx, spaceID, *createHybridBridgesBridge)
	if err != nil {
		return nil, utils.WrapError(err, "failed to create hybrid bridge")
	}
	err = utils.ParseHTTPError(res.Body, res.StatusCode())
	if err != nil {
		return nil, utils.WrapError(err, "failed to parse HTTP error response")
	}

	return res.JSON201, nil
}

// DeleteHybridBridge deletes hybrid bridge.
func DeleteHybridBridge(ctx context.Context, provider *client.NumSpotSDK, hybridBridgesBridgeID string) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.DeleteHybridBridge(ctx, spaceID, utils.StringToUuid(hybridBridgesBridgeID))
	if err != nil {
		return utils.WrapError(err, "failed API client call")
	}
	defer func() { _ = res.Body.Close() }()

	return nil
}

// ReadHybridBridges reads hybrid bridges.
func ReadHybridBridges(ctx context.Context, provider *client.NumSpotSDK) (*[]api.HybridBridge, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := numspotClient.ListHybridBridgesWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list hybrid bridges")
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list hybrid bridges")
	}

	return &read.JSON200.Items, nil
}

// ReadHybridBridge reads hybrid bridge.
func ReadHybridBridge(ctx context.Context, provider *client.NumSpotSDK, hybridBridgesBridgeID string) (*api.HybridBridge, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := numspotClient.ReadHybridBridgeWithResponse(ctx, provider.SpaceID, utils.StringToUuid(hybridBridgesBridgeID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to read hybrid bridge")
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read hybrid bridge")
	}

	return read.JSON200, nil
}
