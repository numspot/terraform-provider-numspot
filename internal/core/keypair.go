package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

// CreateKeypair creates keypair.
func CreateKeypair(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotKeypairCreate *api.CreateKeypairJSONRequestBody,
) (*api.CreateKeypair, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreateResponse *api.CreateKeypairResponse

	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, *numSpotKeypairCreate,
		numspotClient.CreateKeypairWithResponse); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return retryCreateResponse.JSON201, nil
}

// DeleteKeypair deletes keypair.
func DeleteKeypair(ctx context.Context, provider *client.NumSpotSDK, keypairID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, keypairID, numspotClient.DeleteKeypairWithResponse)
	if err != nil {
		return utils.WrapError(err, "failed to delete resource")
	}

	return nil
}

// ReadKeypair reads keypair.
func ReadKeypair(ctx context.Context, provider *client.NumSpotSDK, keypairID string) (*api.Keypair, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadKeypair, err := numspotClient.ReadKeypairsByIdWithResponse(ctx, provider.SpaceID, keypairID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read keypairs by id")
	}
	if err = utils.ParseHTTPError(numSpotReadKeypair.Body, numSpotReadKeypair.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read keypairs by id")
	}

	return (*api.Keypair)(numSpotReadKeypair.JSON200), nil
}

// ReadKeypairs reads keypairs.
func ReadKeypairs(ctx context.Context, provider *client.NumSpotSDK, params *api.ReadKeypairsParams) (numSpotKeypair *[]api.Keypair, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadKeypair, err := numspotClient.ReadKeypairsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read keypairs")
	}
	if err = utils.ParseHTTPError(numSpotReadKeypair.Body, numSpotReadKeypair.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read keypairs")
	}

	return numSpotReadKeypair.JSON200.Items, utils.WrapError(err, "failed to read keypairs")
}
