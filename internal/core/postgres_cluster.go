package core

import (
	"context"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func ReadPostgresClusters(ctx context.Context, provider *client.NumSpotSDK) (*api.PostgresqlListClusters200Response, error) {
	res, err := provider.Client.PostgresqlListClustersWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func CreatePostgresCluster(ctx context.Context, provider *client.NumSpotSDK, body api.PostgresClusterCreationRequest) (*api.PostgresCluster, error) {
	res, err := provider.Client.PostgresqlCreateClusterWithResponse(ctx, provider.SpaceID, body)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	createdID := res.JSON201.Id

	read, err := utils.RetryReadUntilStatusStateValid(ctx, createdID, provider.SpaceID, utils.StateRetryOnCreate, utils.StateStopRetryOnCreate, provider.Client.PostgresqlGetClusterWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotCluster, assert := read.(*api.PostgresCluster)
	if !assert {
		return nil, fmt.Errorf("invalid cluster assertion %s", createdID)
	}

	return numSpotCluster, nil
}

func DeletePostgresCluster(ctx context.Context, provider *client.NumSpotSDK, clusterID api.PostgresClusterIdParameter) (err error) {
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, clusterID, provider.Client.PostgresqlDeleteClusterWithResponse)
}

func ReadPostgresCluster(ctx context.Context, provider *client.NumSpotSDK, paramID api.PostgresClusterIdParameter) (*api.PostgresCluster, error) {
	res, err := provider.Client.PostgresqlGetClusterWithResponse(ctx, provider.SpaceID, paramID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200, nil
}
