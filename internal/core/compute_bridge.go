package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateComputeBridge(ctx context.Context, provider *client.NumSpotSDK, createComputeBridge api.CreateComputeBridgeRequest) (*api.CreateComputeBridge201Response, error) {
	numsClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numsClient.CreateComputeBridgeWithResponse(ctx, provider.SpaceID, createComputeBridge)
	if err != nil {
		return nil, err
	}

	err = utils.ParseHTTPError(res.Body, res.StatusCode())
	if err != nil {
		return nil, err
	}

	return res.JSON201, nil
}

func DeleteComputeBridge(ctx context.Context, provider *client.NumSpotSDK, computeBridgeID api.ResourceIdentifier) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	res, err := numspotClient.DeleteComputeBridgeWithResponse(ctx, spaceID, computeBridgeID)
	if err != nil {
		return err
	}
	err = utils.ParseHTTPError(res.Body, res.StatusCode())
	if err != nil {
		return err
	}

	return nil
}

func ReadComputeBridges(ctx context.Context, provider *client.NumSpotSDK) (*api.ComputeBridges, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := numspotClient.ListComputeBridgesWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}

func ReadComputeBridge(ctx context.Context, provider *client.NumSpotSDK, computeBridgeID api.ResourceIdentifier) (*api.ComputeBridge, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := numspotClient.ReadComputeBridgeWithResponse(ctx, provider.SpaceID, computeBridgeID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}
