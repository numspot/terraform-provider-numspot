package core

import (
	"context"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	clientGatewayPendingStates = []string{pending}
	clientGatewayTargetStates  = []string{available}
)

func CreateClientGateway(ctx context.Context, provider *client.NumSpotSDK, numSpotClientGatewayCreate api.CreateClientGatewayJSONRequestBody) (numSpotClientGateway *api.ClientGateway, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *api.CreateClientGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotClientGatewayCreate, numspotClient.CreateClientGatewayWithResponse); err != nil {
		return nil, err
	}

	createdId := retryCreate.JSON201.Id

	return RetryReadClientGateway(ctx, provider, createOp, createdId)
}

func DeleteClientGateway(ctx context.Context, provider *client.NumSpotSDK, clientGatewayID api.ResourceIdentifier) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, clientGatewayID, numspotClient.DeleteClientGatewayWithResponse)
}

func ReadClientGateway(ctx context.Context, provider *client.NumSpotSDK, clientGatewayID api.ResourceIdentifier) (*api.ClientGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotClientGateway, err := numspotClient.ReadClientGatewayWithResponse(ctx, provider.SpaceID, clientGatewayID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotClientGateway.Body, numSpotClientGateway.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotClientGateway.JSON200, nil
}

func ReadClientGateways(ctx context.Context, provider *client.NumSpotSDK) ([]api.ClientGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotClientGateways, err := numspotClient.ListClientGatewaysWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotClientGateways.Body, numSpotClientGateways.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotClientGateways.JSON200.Items, nil
}

func RetryReadClientGateway(ctx context.Context, provider *client.NumSpotSDK, op string, clientGatewayID api.ResourceIdentifier) (*api.ClientGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := utils.RetryReadUntilStateValid(ctx, clientGatewayID, provider.SpaceID, clientGatewayPendingStates, clientGatewayTargetStates, numspotClient.ReadClientGatewayWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotClientGateway, assert := read.(*api.ClientGateway)
	if !assert {
		return nil, fmt.Errorf("invalid client gateway assertion %s: %s", clientGatewayID, op)
	}
	return numSpotClientGateway, nil
}
