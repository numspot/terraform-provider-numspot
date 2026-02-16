package core

import (
	"context"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func ReadKubernetesClusters(ctx context.Context, provider *client.NumSpotSDK) ([]api.KubernetesCluster, error) {
	res, err := provider.Client.ListKubernetesClustersWithResponse(ctx, provider.SpaceID, nil)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200.Items, err
}

func CreateKubernetesCluster(ctx context.Context, provider *client.NumSpotSDK, numSpotClusterCreate api.CreateKubernetesClusterJSONRequestBody) (*api.KubernetesCluster, error) {
	res, err := provider.Client.CreateKubernetesClusterWithResponse(ctx, provider.SpaceID, numSpotClusterCreate)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	createdID := res.JSON201.Id

	read, err := utils.RetryReadUntilStatusStateValid(ctx, createdID, provider.SpaceID, utils.StateRetryOnCreate, utils.StateStopRetryOnCreate, provider.Client.GetKubernetesClusterWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotCluster, assert := read.(*api.KubernetesCluster)
	if !assert {
		return nil, fmt.Errorf("invalid cluster assertion %s", createdID)
	}

	return numSpotCluster, nil
}

func ReadKubernetesCluster(ctx context.Context, provider *client.NumSpotSDK, clusterId api.ClusterId) (*api.GetKubernetesCluster200Response, error) {
	res, err := provider.Client.GetKubernetesClusterWithResponse(ctx, provider.SpaceID, clusterId)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func DeleteKubernetesCluster(ctx context.Context, provider *client.NumSpotSDK, clusterId api.ClusterId) (err error) {
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, clusterId, provider.Client.DeleteKubernetesClusterWithResponse)
}
