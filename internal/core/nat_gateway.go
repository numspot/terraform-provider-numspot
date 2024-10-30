package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	natGatewayPendingStates = []string{pending}
	natGatewayTargetStates  = []string{available}
)

func CreateNatGateway(ctx context.Context, provider *client.NumSpotSDK, tags []numspot.ResourceTag, body numspot.CreateNatGatewayJSONRequestBody) (numSpotNatGateway *numspot.NatGateway, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreateNatGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, body, numspotClient.CreateNatGatewayWithResponse); err != nil {
		return nil, err
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, createdId, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadNatGateway(ctx, provider, createdId)
}

func UpdateNatGatewayTags(ctx context.Context, provider *client.NumSpotSDK, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, natGatewayID string) (*numspot.NatGateway, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, natGatewayID); err != nil {
		return nil, err
	}
	return ReadNatGateway(ctx, provider, natGatewayID)
}

func DeleteNatGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, natGatewayID, numspotClient.DeleteNatGatewayWithResponse)
	if err != nil {
		return err
	}
	return nil
}

func ReadNatGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) (*numspot.NatGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotNatGateway, err := numspotClient.ReadNatGatewayByIdWithResponse(ctx, provider.SpaceID, natGatewayID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotNatGateway.Body, numSpotNatGateway.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotNatGateway.JSON200, nil
}

func RetryReadNatGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) (*numspot.NatGateway, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, natGatewayID, provider.SpaceID, natGatewayPendingStates, natGatewayTargetStates,
		numspotClient.ReadNatGatewayByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotNatGateway, assert := read.(*numspot.NatGateway)
	if !assert {
		return nil, fmt.Errorf("invalid nat gateway assertion %s", natGatewayID)
	}
	return numSpotNatGateway, nil
}

func ReadNatGatewaysWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadNatGatewayParams) (numSpotNatGateway *[]numspot.NatGateway, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadNatGateway, err := numspotClient.ReadNatGatewayWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadNatGateway.Body, numSpotReadNatGateway.StatusCode()); err != nil {
		return nil, err
	}
	if numSpotReadNatGateway.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of public Ips but got nil")
	}

	return numSpotReadNatGateway.JSON200.Items, err
}
