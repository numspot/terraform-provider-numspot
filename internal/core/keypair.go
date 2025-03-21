package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateKeypair(ctx context.Context, provider *client.NumSpotSDK, numSpotKeypairCreate api.CreateKeypairJSONRequestBody) (*api.CreateKeypair, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreateResponse *api.CreateKeypairResponse

	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotKeypairCreate,
		numspotClient.CreateKeypairWithResponse); err != nil {
		return nil, err
	}

	return retryCreateResponse.JSON201, nil
}

func DeleteKeypair(ctx context.Context, provider *client.NumSpotSDK, keypairID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, keypairID, numspotClient.DeleteKeypairWithResponse)
	if err != nil {
		return err
	}

	return nil
}

func ReadKeypair(ctx context.Context, provider *client.NumSpotSDK, keypairID string) (*api.Keypair, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadKeypair, err := numspotClient.ReadKeypairsByIdWithResponse(ctx, provider.SpaceID, keypairID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadKeypair.Body, numSpotReadKeypair.StatusCode()); err != nil {
		return nil, err
	}

	return (*api.Keypair)(numSpotReadKeypair.JSON200), nil
}

func ReadKeypairs(ctx context.Context, provider *client.NumSpotSDK, params api.ReadKeypairsParams) (numSpotKeypair *[]api.Keypair, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadKeypair, err := numspotClient.ReadKeypairsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadKeypair.Body, numSpotReadKeypair.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadKeypair.JSON200.Items, err
}
