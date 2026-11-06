package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// errInvalidPostgresClusterAssertion is returned when a PostgreSQL cluster read
// result cannot be asserted to the expected type.
var errInvalidPostgresClusterAssertion = errors.New("invalid cluster assertion")

// ReadPostgresClusters reads postgres clusters.
func ReadPostgresClusters(ctx context.Context, provider *client.NumspotSDK) (*[]api.PostgreSQLCluster, error) {
	res, err := provider.Client.PostgreSQLListClustersWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to postgre sql list clusters")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to postgre sql list clusters")
	}

	return &res.JSON200.Items, nil
}

// CreatePostgresCluster creates postgres cluster.
func CreatePostgresCluster(
	ctx context.Context,
	provider *client.NumspotSDK,
	body *api.PostgreSQLClusterCreationRequest,
) (*api.PostgreSQLCluster, error) {
	res, err := provider.Client.PostgreSQLCreateClusterWithResponse(ctx, provider.SpaceID, *body)
	if err != nil {
		return nil, utils.WrapError(err, "failed to postgre sql create cluster")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to postgre sql create cluster")
	}

	if res.JSON201 == nil {
		return nil, fmt.Errorf("%w (HTTP %d)", errMissingCreateID, res.StatusCode())
	}

	createdID := res.JSON201.Id

	read, err := utils.RetryReadUntilStatusStateValid(
		ctx,
		createdID,
		provider.SpaceID,
		utils.StateRetryOnCreate,
		utils.StateStopRetryOnCreate,
		provider.Client.PostgreSQLGetClusterWithResponse,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numspotCluster, assert := read.(*api.PostgreSQLCluster)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidPostgresClusterAssertion, createdID)
	}

	return numspotCluster, nil
}

// DeletePostgresCluster deletes postgres cluster.
func DeletePostgresCluster(ctx context.Context, provider *client.NumspotSDK, clusterID string) (err error) {
	err = utils.RetryDeleteUntilResourceAvailable(
		ctx,
		provider.SpaceID,
		utils.StringToUuid(clusterID),
		provider.Client.PostgreSQLDeleteClusterWithResponse,
	)
	return utils.WrapError(err, "failed to delete resource")
}

// ReadPostgresCluster reads postgres cluster.
func ReadPostgresCluster(ctx context.Context, provider *client.NumspotSDK, clusterID string) (*api.PostgreSQLCluster, error) {
	res, err := provider.Client.PostgreSQLGetClusterWithResponse(ctx, provider.SpaceID, utils.StringToUuid(clusterID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to postgre sql get cluster")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to postgre sql get cluster")
	}

	return res.JSON200, nil
}
