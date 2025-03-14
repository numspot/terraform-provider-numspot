package core

import (
	"context"
	"fmt"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

var (
	virtualGatewayPendingStates = []string{pending}
	virtualGatewayTargetStates  = []string{available}
)

func CreateVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, numSpotVirtualGatewayCreate numspot.CreateVirtualGatewayJSONRequestBody, vpcId string, tags []numspot.ResourceTag) (numSpotVirtualGateway *numspot.VirtualGateway, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreateVirtualGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotVirtualGatewayCreate, numspotClient.CreateVirtualGatewayWithResponse); err != nil {
		return nil, err
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdId, tags); err != nil {
			return nil, err
		}
	}

	// Link virtual gateway to a VPC
	if vpcId != "" {
		err := linkVirtualGateway(ctx, provider, createdId, vpcId)
		if err != nil {
			return nil, err
		}
	}

	return RetryReadVirtualGateway(ctx, provider, createOp, createdId)
}

func linkVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, gatewayId, vpcId string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	_, err = numspotClient.LinkVirtualGatewayToVpcWithResponse(
		ctx,
		provider.SpaceID,
		gatewayId,
		numspot.LinkVirtualGatewayToVpcJSONRequestBody{
			VpcId: vpcId,
		},
	)
	return err
}

func unlinkVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, gatewayId, vpcId string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	_, err = numspotClient.UnlinkVirtualGatewayToVpcWithResponse(
		ctx,
		provider.SpaceID,
		gatewayId,
		numspot.UnlinkVirtualGatewayToVpcJSONRequestBody{
			VpcId: vpcId,
		},
	)
	return err
}

func UpdateVirtualGatewayTags(ctx context.Context, provider *client.NumSpotSDK, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, virtualGatewayID string) (*numspot.VirtualGateway, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, virtualGatewayID); err != nil {
		return nil, err
	}
	return RetryReadVirtualGateway(ctx, provider, updateOp, virtualGatewayID)
}

func DeleteVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, virtualGatewayID, vpcId string) error {
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

func ReadVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, virtualGatewayID string) (*numspot.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotVirtualGateway, err := numspotClient.ReadVirtualGatewaysByIdWithResponse(ctx, provider.SpaceID, virtualGatewayID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotVirtualGateway.Body, numSpotVirtualGateway.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotVirtualGateway.JSON200, nil
}

func ReadVirtualGatewaysWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadVirtualGatewaysParams) (*[]numspot.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotVirtualGateway, err := numspotClient.ReadVirtualGatewaysWithResponse(ctx, provider.SpaceID, &params)
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

func RetryReadVirtualGateway(ctx context.Context, provider *client.NumSpotSDK, op string, virtualGatewayID string) (*numspot.VirtualGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, virtualGatewayID, provider.SpaceID, virtualGatewayPendingStates, virtualGatewayTargetStates,
		numspotClient.ReadVirtualGatewaysByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotVirtualGateway, assert := read.(*numspot.VirtualGateway)
	if !assert {
		return nil, fmt.Errorf("invalid virtual gateway assertion %s: %s", virtualGatewayID, op)
	}
	return numSpotVirtualGateway, nil
}
