package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	natGatewayPendingStates = []string{pending, deleting}
	natGatewayTargetStates  = []string{available, deleted}
)

func CreateNATGateway(ctx context.Context, provider *client.NumSpotSDK, tags []numspot.ResourceTag, body numspot.CreateNatGatewayJSONRequestBody) (numSpotNatGateway *numspot.NatGateway, err error) {
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
		if err = createTags(ctx, provider, createdId, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadNATGateway(ctx, provider, createdId)
}

func UpdateNATGatewayTags(ctx context.Context, provider *client.NumSpotSDK, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, natGatewayID string) (*numspot.NatGateway, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, natGatewayID); err != nil {
		return nil, err
	}

	return ReadNATGateway(ctx, provider, natGatewayID)
}

func DeleteNATGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	return utils.RetryDeleteUntilResourceAvailable(ctx, spaceID, natGatewayID, numspotClient.DeleteNatGatewayWithResponse)
}

func ReadNATGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) (*numspot.NatGateway, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	numSpotNatGateway, err := numspotClient.ReadNatGatewayByIdWithResponse(ctx, spaceID, natGatewayID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotNatGateway.Body, numSpotNatGateway.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotNatGateway.JSON200, nil
}

func RetryReadNATGateway(ctx context.Context, provider *client.NumSpotSDK, natGatewayID string) (*numspot.NatGateway, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := utils.RetryReadUntilStateValid(ctx, natGatewayID, spaceID, natGatewayPendingStates, natGatewayTargetStates,
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

func ReadNATGatewaysWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadNatGatewayParams) (numSpotNatGateway *[]numspot.NatGateway, err error) {
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
