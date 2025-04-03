package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateManagedServiceBridge(ctx context.Context, provider *client.NumSpotSDK, createManagedServicesBridge api.CreateManagedServicesBridgeRequest) (*api.CreateManagedServicesBridge201Response, error) {
	spaceID := provider.SpaceID

	numsClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numsClient.CreateManagedServicesBridgeWithResponse(ctx, spaceID, createManagedServicesBridge)
	if err != nil {
		return nil, err
	}
	err = utils.ParseHTTPError(res.Body, res.StatusCode())
	if err != nil {
		return nil, err
	}

	return res.JSON201, nil
}

func DeleteManagedServiceBridge(ctx context.Context, provider *client.NumSpotSDK, managedServicesBridgeID api.ResourceIdentifier) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	_, err = numspotClient.DeleteManagedServicesBridge(ctx, spaceID, managedServicesBridgeID)
	if err != nil {
		return err
	}

	return nil
}

func ReadManagedServiceBridges(ctx context.Context, provider *client.NumSpotSDK) (*api.ManagedServicesBridges, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := numspotClient.ListManagedServicesBridgesWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}

func ReadManagedServiceBridge(ctx context.Context, provider *client.NumSpotSDK, managedServicesBridgeID api.ResourceIdentifier) (*api.ManagedServicesBridge, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := numspotClient.ReadManagedServicesBridgeWithResponse(ctx, provider.SpaceID, managedServicesBridgeID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}
