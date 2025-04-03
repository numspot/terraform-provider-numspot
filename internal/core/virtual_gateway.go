package core

import (
	"context"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	virtualGatewayPendingStates = []string{pending}
	virtualGatewayTargetStates  = []string{available}
)

func CreateVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, numSpotVirtualGatewayCreate api.CreateVirtualGatewayJSONRequestBody, vpcId string) (numSpotVirtualGateway *api.VirtualGateway, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *api.CreateVirtualGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotVirtualGatewayCreate, numspotClient.CreateVirtualGatewayWithResponse); err != nil {
		return nil, err
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
		return err
	}

	_, err = numspotClient.LinkVirtualGatewayWithResponse(
		ctx,
		provider.SpaceID,
		gatewayId,
		api.LinkVirtualGatewayJSONRequestBody{
			VpcId: vpcId,
		},
	)
	return err
}

func unlinkVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, gatewayId api.ResourceIdentifier, vpcId string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	_, err = numspotClient.UnlinkVirtualGatewayWithResponse(
		ctx,
		provider.SpaceID,
		gatewayId,
		api.UnlinkVirtualGatewayJSONRequestBody{
			VpcId: vpcId,
		},
	)
	return err
}

func DeleteVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, virtualGatewayID api.ResourceIdentifier, vpcId string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	// Unlink virtual gateway from VPC
	if vpcId != "" {
		err := unlinkVirtualGateway(ctx, provider, virtualGatewayID, vpcId)
		if err != nil {
			return err
		}
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, virtualGatewayID, numspotClient.DeleteVirtualGatewayWithResponse)
	if err != nil {
		return err
	}
	return nil
}

func ReadVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, virtualGatewayID api.ResourceIdentifier) (*api.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotVirtualGateway, err := numspotClient.ReadVirtualGatewayWithResponse(ctx, provider.SpaceID, virtualGatewayID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotVirtualGateway.Body, numSpotVirtualGateway.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotVirtualGateway.JSON200, nil
}

func ReadVirtualGatewaysWithParams(ctx context.Context, provider *client.NumSpotSDK) ([]api.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotVirtualGateway, err := numspotClient.ListVirtualGatewaysWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotVirtualGateway.Body, numSpotVirtualGateway.StatusCode()); err != nil {
		return nil, err
	}

	if numSpotVirtualGateway.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of virtual gateway but got nil")
	}

	return numSpotVirtualGateway.JSON200.Items, nil
}

func RetryReadVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, op string, virtualGatewayID api.ResourceIdentifier) (*api.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, virtualGatewayID, provider.SpaceID, virtualGatewayPendingStates, virtualGatewayTargetStates, numspotClient.ReadVirtualGatewayWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotVirtualGateway, assert := read.(*api.VirtualGateway)
	if !assert {
		return nil, fmt.Errorf("invalid virtual gateway assertion %s: %s", virtualGatewayID, op)
	}
	return numSpotVirtualGateway, nil
}
