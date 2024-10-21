package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateKeypair(ctx context.Context, provider *client.NumSpotSDK, numSpotKeypairCreate numspot.CreateKeypairJSONRequestBody) (*numspot.CreateKeypair, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreateResponse *numspot.CreateKeypairResponse

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

func ReadKeypairsWithID(ctx context.Context, provider *client.NumSpotSDK, keypairID string) (numSpotKeypair *numspot.Keypair, err error) {
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

	keypair := (*numspot.Keypair)(numSpotReadKeypair.JSON200)
	if keypair == nil {
		return nil, fmt.Errorf("Failed to cast read response to keypair object.")
	}
	return keypair, nil
}

func ReadKeypairsWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadKeypairsParams) (numSpotKeypair *[]numspot.Keypair, err error) {
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
	if numSpotReadKeypair.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of keypair but got nil")
	}

	return numSpotReadKeypair.JSON200.Items, err
}
