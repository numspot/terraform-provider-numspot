package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	clientGatewayPendingStates = []string{pending}
	clientGatewayTargetStates  = []string{available}
)

func CreateClientGateway(ctx context.Context, provider services.IProvider, numSpotClientGatewayCreate numspot.CreateClientGatewayJSONRequestBody, tags []numspot.ResourceTag) (numSpotClientGateway *numspot.ClientGateway, err error) {
	spaceID := provider.GetSpaceID()

	var retryCreate *numspot.CreateClientGatewayResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotClientGatewayCreate,
		provider.GetNumspotClient().CreateClientGatewayWithResponse); err != nil {
		return nil, err
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, createdId, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadClientGateway(ctx, provider, createOp, createdId)
}

func UpdateClientGatewayTags(ctx context.Context, provider services.IProvider, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, clientGatewayID string) (*numspot.ClientGateway, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, clientGatewayID); err != nil {
		return nil, err
	}
	return RetryReadClientGateway(ctx, provider, updateOp, clientGatewayID)
}

func DeleteClientGateway(ctx context.Context, provider services.IProvider, clientGatewayID string) error {
	err := utils.RetryDeleteUntilResourceAvailable(ctx, provider.GetSpaceID(), clientGatewayID, provider.GetNumspotClient().DeleteClientGatewayWithResponse)
	if err != nil {
		return err
	}
	return nil
}

func ReadClientGateway(ctx context.Context, provider services.IProvider, clientGatewayID string) (*numspot.ClientGateway, error) {
	numSpotClientGateway, err := provider.GetNumspotClient().ReadClientGatewaysByIdWithResponse(ctx, provider.GetSpaceID(), clientGatewayID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotClientGateway.Body, numSpotClientGateway.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotClientGateway.JSON200, nil
}

func RetryReadClientGateway(ctx context.Context, provider services.IProvider, op string, clientGatewayID string) (*numspot.ClientGateway, error) {
	read, err := utils.RetryReadUntilStateValid(ctx, clientGatewayID, provider.GetSpaceID(), clientGatewayPendingStates, clientGatewayTargetStates,
		provider.GetNumspotClient().ReadClientGatewaysByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotClientGateway, assert := read.(*numspot.ClientGateway)
	if !assert {
		return nil, fmt.Errorf("invalid client gateway assertion %s: %s", clientGatewayID, op)
	}
	return numSpotClientGateway, nil
}
