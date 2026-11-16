package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
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
	provider *client.NumspotSDK,
	numspotClientGatewayCreate *api.CreateClientGatewayJSONRequestBody,
) (numspotClientGateway *api.ClientGateway, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateClientGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*numspotClientGatewayCreate,
		numspotClient.CreateClientGatewayWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if retryCreate.JSON201 == nil {
		return nil, fmt.Errorf("%w (HTTP %d)", errMissingCreateID, retryCreate.StatusCode())
	}

	createdId := retryCreate.JSON201.Id

	return RetryReadClientGateway(ctx, provider, createOp, createdId)
}

// DeleteClientGateway deletes client gateway.
func DeleteClientGateway(ctx context.Context, provider *client.NumspotSDK, clientGatewayID string) error {
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
func ReadClientGateway(ctx context.Context, provider *client.NumspotSDK, clientGatewayID string) (*api.ClientGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotClientGateway, err := numspotClient.ReadClientGatewayWithResponse(ctx, provider.SpaceID, utils.StringToUuid(clientGatewayID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to read client gateway")
	}

	if err = utils.ParseHTTPError(numspotClientGateway.Body, numspotClientGateway.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read client gateway")
	}

	return numspotClientGateway.JSON200, nil
}

// ReadClientGateways reads client gateways.
func ReadClientGateways(ctx context.Context, provider *client.NumspotSDK) (*[]api.ClientGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotClientGateways, err := numspotClient.ListClientGatewaysWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list client gateways")
	}

	if err = utils.ParseHTTPError(numspotClientGateways.Body, numspotClientGateways.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list client gateways")
	}

	return &numspotClientGateways.JSON200.Items, nil
}

// RetryReadClientGateway retries read client gateway.
func RetryReadClientGateway(
	ctx context.Context,
	provider *client.NumspotSDK,
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

	numspotClientGateway, assert := read.(*api.ClientGateway)
	if !assert {
		return nil, fmt.Errorf("%w %s: %s", errInvalidClientGatewayAssertion, clientGatewayID, op)
	}
	return numspotClientGateway, nil
}
