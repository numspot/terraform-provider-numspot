package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

// errInvalidKubernetesClusterAssertion is returned when a kubernetes cluster
// read result cannot be asserted to the expected type.
var errInvalidKubernetesClusterAssertion = errors.New("invalid cluster assertion")

// ReadKubernetesClusters reads kubernetes clusters.
func ReadKubernetesClusters(ctx context.Context, provider *client.NumSpotSDK) (*[]api.KubernetesCluster, error) {
	res, err := provider.Client.ListKubernetesClustersWithResponse(ctx, provider.SpaceID, nil)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list kubernetes clusters")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list kubernetes clusters")
	}

	return &res.JSON200.Items, nil
}

// CreateKubernetesCluster creates kubernetes cluster.
func CreateKubernetesCluster(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotClusterCreate *api.CreateKubernetesClusterJSONRequestBody,
	timeout *time.Duration,
) (*api.KubernetesCluster, error) {
	res, err := provider.Client.CreateKubernetesClusterWithResponse(ctx, provider.SpaceID, *numSpotClusterCreate)
	if err != nil {
		return nil, utils.WrapError(err, "failed to create kubernetes cluster")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to create kubernetes cluster")
	}

	createdID := res.JSON201.Id

	retryTimeout := utils.TfRequestStateRetryTimeout
	if timeout != nil {
		retryTimeout = *timeout
	}

	read, err := utils.RetryReadUntilStatusStateValidWithTimeout(
		ctx,
		createdID,
		provider.SpaceID,
		utils.StateRetryOnCreate,
		utils.StateStopRetryOnCreate,
		provider.Client.GetKubernetesClusterWithResponse,
		retryTimeout,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numSpotCluster, assert := read.(*api.KubernetesCluster)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidKubernetesClusterAssertion, createdID)
	}

	return numSpotCluster, nil
}

// ReadKubernetesCluster reads kubernetes cluster.
func ReadKubernetesCluster(ctx context.Context, provider *client.NumSpotSDK, clusterID string) (*api.KubernetesCluster, error) {
	res, err := provider.Client.GetKubernetesClusterWithResponse(ctx, provider.SpaceID, utils.StringToUuid(clusterID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to get kubernetes cluster")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get kubernetes cluster")
	}

	return res.JSON200, nil
}

// DeleteKubernetesCluster deletes kubernetes cluster.
func DeleteKubernetesCluster(ctx context.Context, provider *client.NumSpotSDK, clusterID string) (err error) {
	err = utils.RetryDeleteUntilResourceAvailable(
		ctx,
		provider.SpaceID,
		utils.StringToUuid(clusterID),
		provider.Client.DeleteKubernetesClusterWithResponse,
	)
	return utils.WrapError(err, "failed to delete resource")
}
