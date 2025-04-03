package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateHybridBridge(ctx context.Context, provider *client.NumSpotSDK, createHybridBridgesBridge api.CreateHybridBridgeRequest) (*api.CreateHybridBridge201Response, error) {
	spaceID := provider.SpaceID

	numsClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numsClient.CreateHybridBridgeWithResponse(ctx, spaceID, createHybridBridgesBridge)
	if err != nil {
		return nil, err
	}
	err = utils.ParseHTTPError(res.Body, res.StatusCode())
	if err != nil {
		return nil, err
	}

	return res.JSON201, nil
}

func DeleteHybridBridge(ctx context.Context, provider *client.NumSpotSDK, hybridBridgesBridgeID api.ResourceIdentifier) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	_, err = numspotClient.DeleteHybridBridge(ctx, spaceID, hybridBridgesBridgeID)
	if err != nil {
		return err
	}

	return nil
}

func ReadHybridBridges(ctx context.Context, provider *client.NumSpotSDK) (*api.HybridBridges, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := numspotClient.ListHybridBridgesWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}

func ReadHybridBridge(ctx context.Context, provider *client.NumSpotSDK, HybridBridgesBridgeID api.ResourceIdentifier) (*api.HybridBridge, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := numspotClient.ReadHybridBridgeWithResponse(ctx, provider.SpaceID, HybridBridgesBridgeID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}
