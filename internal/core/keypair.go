package core

import (
	"context"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// CreateKeypair creates keypair.
func CreateKeypair(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotKeypairCreate *api.CreateKeypairJSONRequestBody,
) (*api.CreateKeypair, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreateResponse *api.CreateKeypairResponse

	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, *numspotKeypairCreate,
		numspotClient.CreateKeypairWithResponse); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return retryCreateResponse.JSON201, nil
}

// DeleteKeypair deletes keypair.
func DeleteKeypair(ctx context.Context, provider *client.NumspotSDK, keypairID string) (err error) {
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
func ReadKeypair(ctx context.Context, provider *client.NumspotSDK, keypairID string) (*api.Keypair, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadKeypair, err := numspotClient.ReadKeypairsByIdWithResponse(ctx, provider.SpaceID, keypairID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read keypairs by id")
	}
	if err = utils.ParseHTTPError(numspotReadKeypair.Body, numspotReadKeypair.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read keypairs by id")
	}

	return (*api.Keypair)(numspotReadKeypair.JSON200), nil
}

// ReadKeypairs reads keypairs.
func ReadKeypairs(ctx context.Context, provider *client.NumspotSDK, params *api.ReadKeypairsParams) (numspotKeypair *[]api.Keypair, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadKeypair, err := numspotClient.ReadKeypairsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read keypairs")
	}
	if err = utils.ParseHTTPError(numspotReadKeypair.Body, numspotReadKeypair.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read keypairs")
	}

	return numspotReadKeypair.JSON200.Items, utils.WrapError(err, "failed to read keypairs")
}
