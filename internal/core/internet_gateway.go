package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateInternetGateway(ctx context.Context, provider *client.NumSpotSDK, tags []numspot.ResourceTag, vpcID string) (*numspot.InternetGateway, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreateResponse *numspot.CreateInternetGatewayResponse
	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailable(ctx, spaceID, numspotClient.CreateInternetGatewayWithResponse); err != nil {
		return nil, err
	}

	internetGatewayID := *retryCreateResponse.JSON201.Id

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, internetGatewayID, tags); err != nil {
			return nil, err
		}
	}

	if vpcID != "" {
		linkResponse, err := numspotClient.LinkInternetGatewayWithResponse(ctx, provider.SpaceID, internetGatewayID,
			numspot.LinkInternetGatewayJSONRequestBody{
				VpcId: vpcID,
			},
		)
		if err != nil {
			return nil, err
		}
		if err = utils.ParseHTTPError(linkResponse.Body, linkResponse.StatusCode()); err != nil {
			return nil, err
		}
	}

	internetGateway, err := ReadInternetGatewaysWithID(ctx, provider, internetGatewayID)
	if err != nil {
		return nil, err
	}

	return internetGateway, nil
}

func UpdateInternetGatewayTags(ctx context.Context, provider *client.NumSpotSDK, internetGatewayID string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.InternetGateway, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, internetGatewayID); err != nil {
		return nil, err
	}
	return ReadInternetGatewaysWithID(ctx, provider, internetGatewayID)
}

func DeleteInternetGateway(ctx context.Context, provider *client.NumSpotSDK, internetGatewayID string, vpcID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	if vpcID != "" {
		_, _ = numspotClient.UnlinkInternetGatewayWithResponse(ctx, provider.SpaceID, internetGatewayID,
			numspot.UnlinkInternetGatewayJSONRequestBody{
				VpcId: vpcID,
			},
		)
		// Error not handled, we try to delete internet gateway anyway
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, internetGatewayID, numspotClient.DeleteInternetGatewayWithResponse)
	if err != nil {
		return err
	}

	return nil
}

func ReadInternetGatewaysWithID(ctx context.Context, provider *client.NumSpotSDK, internetGatewayID string) (numSpotInternetGateway *numspot.InternetGateway, err error) {
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

func ReadInternetGatewaysWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadInternetGatewaysParams) (numSpotInternetGateway *[]numspot.InternetGateway, err error) {
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
	if numSpotReadInternetGateway.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of internetGateway but got nil")
	}

	return numSpotReadInternetGateway.JSON200.Items, err
}
