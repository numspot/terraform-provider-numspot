package core

import (
	"context"

	"github.com/google/uuid"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func ReadKubernetesNodePools(ctx context.Context, provider *client.NumSpotSDK, clusterId api.ClusterId) (*api.ListKubernetesNodePools200Response, error) {
	res, err := provider.Client.ListKubernetesNodePoolsWithResponse(ctx, provider.SpaceID, clusterId, nil)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func CreateKubernetesNodePool(ctx context.Context, provider *client.NumSpotSDK, numSpotNodePoolCreate api.CreateKubernetesNodePoolJSONRequestBody, clusterId api.ClusterId) (*api.CreateKubernetesNodePool201Response, error) {
	res, err := provider.Client.CreateKubernetesNodePoolWithResponse(ctx, provider.SpaceID, clusterId, numSpotNodePoolCreate)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON201, nil
}

func ReadKubernetesNodePool(ctx context.Context, provider *client.NumSpotSDK, clusterId api.ClusterId, nodePoolId string) (*api.GetKubernetesNodePool200Response, error) {
	res, err := provider.Client.GetKubernetesNodePoolWithResponse(ctx, provider.SpaceID, clusterId, uuid.MustParse(nodePoolId))
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func DeleteKubernetesNodePool(ctx context.Context, provider *client.NumSpotSDK, clusterId api.ClusterId, nodePoolId string) (err error) {
	res, err := provider.Client.DeleteKubernetesNodePoolWithResponse(ctx, provider.SpaceID, clusterId, uuid.MustParse(nodePoolId))
	if err != nil {
		return err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return err
	}

	return nil
}
