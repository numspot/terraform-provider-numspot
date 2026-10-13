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
	virtualGatewayPendingStates = []string{pending}
	virtualGatewayTargetStates  = []string{available}

	// errExpectedVirtualGatewayList is returned when a virtual gateway list
	// response is unexpectedly nil.
	errExpectedVirtualGatewayList = errors.New("HTTP call failed : expected a list of virtual gateway but got nil")
	// errInvalidVirtualGatewayAssertion is returned when a virtual gateway read
	// result cannot be asserted to the expected type.
	errInvalidVirtualGatewayAssertion = errors.New("invalid virtual gateway assertion")
)

// CreateVirtualGateway creates virtual gateway.
func CreateVirtualGateway(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotVirtualGatewayCreate *api.CreateVirtualGatewayJSONRequestBody,
	vpcId string,
) (numSpotVirtualGateway *api.VirtualGateway, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateVirtualGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*numSpotVirtualGatewayCreate,
		numspotClient.CreateVirtualGatewayWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	createdId := retryCreate.JSON201.Id

	// Link virtual gateway to a VPC
	if vpcId != "" {
		err := linkVirtualGateway(ctx, provider, createdId, vpcId)
		if err != nil {
			return nil, err
		}
	}

	return RetryReadVirtualGateway(ctx, provider, createOp, createdId)
}

func linkVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, gatewayId api.ResourceIdentifier, vpcId string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	_, err = numspotClient.LinkVirtualGatewayWithResponse(
		ctx,
		provider.SpaceID,
		gatewayId,
		api.LinkVirtualGatewayJSONRequestBody{
			VpcId: vpcId,
		},
	)
	return utils.WrapError(err, "failed to link virtual gateway")
}

func unlinkVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, gatewayId api.ResourceIdentifier, vpcId string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	_, err = numspotClient.UnlinkVirtualGatewayWithResponse(
		ctx,
		provider.SpaceID,
		gatewayId,
		api.UnlinkVirtualGatewayJSONRequestBody{
			VpcId: vpcId,
		},
	)
	return utils.WrapError(err, "failed to unlink virtual gateway")
}

// DeleteVirtualGateway deletes virtual gateway.
func DeleteVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, virtualGatewayID, vpcId string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	gatewayID := utils.StringToUuid(virtualGatewayID)

	// Unlink virtual gateway from VPC
	if vpcId != "" {
		err := unlinkVirtualGateway(ctx, provider, gatewayID, vpcId)
		if err != nil {
			return err
		}
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, gatewayID, numspotClient.DeleteVirtualGatewayWithResponse)
	if err != nil {
		return utils.WrapError(err, "failed to delete resource")
	}
	return nil
}

// ReadVirtualGateway reads virtual gateway.
func ReadVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, virtualGatewayID string) (*api.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotVirtualGateway, err := numspotClient.ReadVirtualGatewayWithResponse(ctx, provider.SpaceID, utils.StringToUuid(virtualGatewayID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to read virtual gateway")
	}

	if err = utils.ParseHTTPError(numSpotVirtualGateway.Body, numSpotVirtualGateway.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read virtual gateway")
	}

	return numSpotVirtualGateway.JSON200, nil
}

// ReadVirtualGatewaysWithParams reads virtual gateways with params.
func ReadVirtualGatewaysWithParams(ctx context.Context, provider *client.NumSpotSDK) (*[]api.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotVirtualGateway, err := numspotClient.ListVirtualGatewaysWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list virtual gateways")
	}

	if err = utils.ParseHTTPError(numSpotVirtualGateway.Body, numSpotVirtualGateway.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list virtual gateways")
	}

	if numSpotVirtualGateway.JSON200.Items == nil {
		return nil, errExpectedVirtualGatewayList
	}

	return &numSpotVirtualGateway.JSON200.Items, nil
}

// RetryReadVirtualGateway retries read virtual gateway.
func RetryReadVirtualGateway(
	ctx context.Context,
	provider *client.NumSpotSDK,
	op string,
	virtualGatewayID api.ResourceIdentifier,
) (*api.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		virtualGatewayID,
		provider.SpaceID,
		virtualGatewayPendingStates,
		virtualGatewayTargetStates,
		numspotClient.ReadVirtualGatewayWithResponse,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numSpotVirtualGateway, assert := read.(*api.VirtualGateway)
	if !assert {
		return nil, fmt.Errorf("%w %s: %s", errInvalidVirtualGatewayAssertion, virtualGatewayID, op)
	}
	return numSpotVirtualGateway, nil
}
