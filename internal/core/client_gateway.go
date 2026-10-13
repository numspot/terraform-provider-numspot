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
	clientGatewayPendingStates = []string{pending}
	clientGatewayTargetStates  = []string{available}

	// errInvalidClientGatewayAssertion is returned when a client gateway does
	// not reach the expected state after creation.
	errInvalidClientGatewayAssertion = errors.New("invalid client gateway assertion")
)

// CreateClientGateway creates client gateway.
func CreateClientGateway(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotClientGatewayCreate *api.CreateClientGatewayJSONRequestBody,
) (numSpotClientGateway *api.ClientGateway, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateClientGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*numSpotClientGatewayCreate,
		numspotClient.CreateClientGatewayWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	createdId := retryCreate.JSON201.Id

	return RetryReadClientGateway(ctx, provider, createOp, createdId)
}

// DeleteClientGateway deletes client gateway.
func DeleteClientGateway(ctx context.Context, provider *client.NumSpotSDK, clientGatewayID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	err = utils.RetryDeleteUntilResourceAvailable(
		ctx,
		provider.SpaceID,
		utils.StringToUuid(clientGatewayID),
		numspotClient.DeleteClientGatewayWithResponse,
	)
	return utils.WrapError(err, "failed to delete resource")
}

// ReadClientGateway reads client gateway.
func ReadClientGateway(ctx context.Context, provider *client.NumSpotSDK, clientGatewayID string) (*api.ClientGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotClientGateway, err := numspotClient.ReadClientGatewayWithResponse(ctx, provider.SpaceID, utils.StringToUuid(clientGatewayID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to read client gateway")
	}

	if err = utils.ParseHTTPError(numSpotClientGateway.Body, numSpotClientGateway.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read client gateway")
	}

	return numSpotClientGateway.JSON200, nil
}

// ReadClientGateways reads client gateways.
func ReadClientGateways(ctx context.Context, provider *client.NumSpotSDK) (*[]api.ClientGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotClientGateways, err := numspotClient.ListClientGatewaysWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list client gateways")
	}

	if err = utils.ParseHTTPError(numSpotClientGateways.Body, numSpotClientGateways.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list client gateways")
	}

	return &numSpotClientGateways.JSON200.Items, nil
}

// RetryReadClientGateway retries read client gateway.
func RetryReadClientGateway(
	ctx context.Context,
	provider *client.NumSpotSDK,
	op string,
	clientGatewayID api.ResourceIdentifier,
) (*api.ClientGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := utils.RetryReadUntilStateValid(
		ctx,
		clientGatewayID,
		provider.SpaceID,
		clientGatewayPendingStates,
		clientGatewayTargetStates,
		numspotClient.ReadClientGatewayWithResponse,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numSpotClientGateway, assert := read.(*api.ClientGateway)
	if !assert {
		return nil, fmt.Errorf("%w %s: %s", errInvalidClientGatewayAssertion, clientGatewayID, op)
	}
	return numSpotClientGateway, nil
}
