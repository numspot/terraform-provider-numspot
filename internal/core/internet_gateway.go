package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateInternetGateway(ctx context.Context, provider *client.NumSpotSDK, tags []api.ResourceTag, vpcID string) (*api.InternetGateway, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreateResponse *api.CreateInternetGatewayResponse
	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailable(ctx, spaceID, numspotClient.CreateInternetGatewayWithResponse); err != nil {
		return nil, err
	}

	internetGatewayID := *retryCreateResponse.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, internetGatewayID, tags); err != nil {
			return nil, err
		}
	}

	if vpcID != "" {
		var linkVPCResponse *api.LinkInternetGatewayResponse
		if linkVPCResponse, err = numspotClient.LinkInternetGatewayWithResponse(ctx, provider.SpaceID, internetGatewayID,
			api.LinkInternetGatewayJSONRequestBody{
				VpcId: vpcID,
			},
		); err != nil {
			return nil, err
		}
		if err = utils.ParseHTTPError(linkVPCResponse.Body, linkVPCResponse.StatusCode()); err != nil {
			return nil, err
		}
	}

	internetGateway, err := ReadInternetGatewaysWithID(ctx, provider, internetGatewayID)
	if err != nil {
		return nil, err
	}

	return internetGateway, nil
}

func UpdateInternetGatewayTags(ctx context.Context, provider *client.NumSpotSDK, internetGatewayID string, stateTags []api.ResourceTag, planTags []api.ResourceTag) (*api.InternetGateway, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, internetGatewayID); err != nil {
		return nil, err
	}
	return ReadInternetGatewaysWithID(ctx, provider, internetGatewayID)
}

func DeleteInternetGateway(ctx context.Context, provider *client.NumSpotSDK, internetGatewayID string, vpcID string) (err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	if vpcID != "" {
		if _, err = utils.RetryUntilResourceAvailableWithBody(ctx, spaceID, internetGatewayID,
			api.UnlinkInternetGatewayJSONRequestBody{
				VpcId: vpcID,
			}, numspotClient.UnlinkInternetGatewayWithResponse); err != nil {
			return err
		}
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, internetGatewayID, numspotClient.DeleteInternetGatewayWithResponse)
	if err != nil {
		return err
	}

	return nil
}

func ReadInternetGatewaysWithID(ctx context.Context, provider *client.NumSpotSDK, internetGatewayID string) (numSpotInternetGateway *api.InternetGateway, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadInternetGateway, err := numspotClient.ReadInternetGatewaysByIdWithResponse(ctx, provider.SpaceID, internetGatewayID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadInternetGateway.Body, numSpotReadInternetGateway.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadInternetGateway.JSON200, err
}

func ReadInternetGatewaysWithParams(ctx context.Context, provider *client.NumSpotSDK, params api.ReadInternetGatewaysParams) (numSpotInternetGateway *[]api.InternetGateway, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadInternetGateway, err := numspotClient.ReadInternetGatewaysWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadInternetGateway.Body, numSpotReadInternetGateway.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadInternetGateway.JSON200.Items, err
}
