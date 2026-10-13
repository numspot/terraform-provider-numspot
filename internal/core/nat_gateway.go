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
	natGatewayPendingStates = []string{pending, deleting}
	natGatewayTargetStates  = []string{available, deleted}

	// errInvalidNatGatewayAssertion is returned when a NAT gateway read result
	// cannot be asserted to the expected type.
	errInvalidNatGatewayAssertion = errors.New("invalid nat gateway assertion")
	// errExpectedNatGatewayList is returned when a NAT gateway list response is
	// unexpectedly nil.
	errExpectedNatGatewayList = errors.New("HTTP call failed : expected a list of public Ips but got nil")
)

// CreateNatGateway creates NAT gateway.
func CreateNatGateway(
	ctx context.Context,
	provider *client.NumSpotSDK,
	body *api.CreateNatGatewayJSONRequestBody,
	tags []api.ResourceTag,
) (numSpotNatGateway *api.NatGateway, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateNatGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*body,
		numspotClient.CreateNatGatewayWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdId, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	return RetryReadNatGateway(ctx, provider, createdId)
}

// UpdateNatGatewayTags updates NAT gateway tags.
func UpdateNatGatewayTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	natGatewayID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.NatGateway, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, natGatewayID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return ReadNatGateway(ctx, provider, natGatewayID)
}

// DeleteNatGateway deletes NAT gateway.
func DeleteNatGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, spaceID, natGatewayID, numspotClient.DeleteNatGatewayWithResponse),
		"failed to delete resource",
	)
}

// ReadNatGateway reads NAT gateway.
func ReadNatGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) (*api.NatGateway, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numSpotNatGateway, err := numspotClient.ReadNatGatewayByIdWithResponse(ctx, spaceID, natGatewayID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read nat gateway by id")
	}

	if err = utils.ParseHTTPError(numSpotNatGateway.Body, numSpotNatGateway.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read nat gateway by id")
	}

	return numSpotNatGateway.JSON200, nil
}

// RetryReadNatGateway retries read NAT gateway.
func RetryReadNatGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) (*api.NatGateway, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := utils.RetryReadUntilStateValid(ctx, natGatewayID, spaceID, natGatewayPendingStates, natGatewayTargetStates,
		numspotClient.ReadNatGatewayByIdWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numSpotNatGateway, assert := read.(*api.NatGateway)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidNatGatewayAssertion, natGatewayID)
	}
	return numSpotNatGateway, nil
}

// ReadNatGatewaysWithParams reads NAT gateways with params.
func ReadNatGatewaysWithParams(
	ctx context.Context,
	provider *client.NumSpotSDK,
	params *api.ReadNatGatewayParams,
) (numSpotNatGateway *[]api.NatGateway, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadNatGateway, err := numspotClient.ReadNatGatewayWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read nat gateway")
	}
	if err = utils.ParseHTTPError(numSpotReadNatGateway.Body, numSpotReadNatGateway.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read nat gateway")
	}
	if numSpotReadNatGateway.JSON200.Items == nil {
		return nil, errExpectedNatGatewayList
	}

	return numSpotReadNatGateway.JSON200.Items, utils.WrapError(err, "failed to read nat gateway")
}
