package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// kubernetesNodePoolCreateRetryTimeout is the default timeout for waiting until
// a created node pool reaches a terminal state. Node pools provision their
// underlying machines for up to ~1h, so the create call blocks until the pool
// is RUNNING instead of returning the 201 payload while still provisioning.
const kubernetesNodePoolCreateRetryTimeout = time.Hour

// errInvalidKubernetesNodePoolAssertion is returned when a kubernetes node pool
// read result cannot be asserted to the expected type.
var errInvalidKubernetesNodePoolAssertion = errors.New("invalid node pool assertion")

// ReadKubernetesNodePools reads kubernetes node pools.
func ReadKubernetesNodePools(ctx context.Context, provider *client.NumspotSDK, clusterId string) (*[]api.KubernetesNodePool, error) {
	parsedClusterID, err := uuid.Parse(clusterId)
	if err != nil {
		return nil, utils.WrapError(err, "failed to parse UUID")
	}

	res, err := provider.Client.ListKubernetesNodePoolsWithResponse(ctx, provider.SpaceID, parsedClusterID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list kubernetes node pools")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list kubernetes node pools")
	}

	return &res.JSON200.Items, nil
}

// CreateKubernetesNodePool creates kubernetes node pool and waits until the
// pool reaches a terminal state (RUNNING/FAILED), since provisioning takes up
// to ~1h. The optional timeout overrides the default 1h retry window.
func CreateKubernetesNodePool(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotNodePoolCreate *api.CreateKubernetesNodePoolJSONRequestBody,
	clusterId string,
	timeout *time.Duration,
) (*api.KubernetesNodePool, error) {
	parsedClusterID, err := uuid.Parse(clusterId)
	if err != nil {
		return nil, utils.WrapError(err, "failed to parse UUID")
	}

	res, err := provider.Client.CreateKubernetesNodePoolWithResponse(ctx, provider.SpaceID, parsedClusterID, *numspotNodePoolCreate)
	if err != nil {
		return nil, utils.WrapError(err, "failed to create kubernetes node pool")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to create kubernetes node pool")
	}

	createdID := res.JSON201.Id

	retryTimeout := kubernetesNodePoolCreateRetryTimeout
	if timeout != nil {
		retryTimeout = *timeout
	}

	read, err := utils.RetryReadUntilStatusStateValidWithTimeout(
		ctx,
		createdID,
		provider.SpaceID,
		utils.StateRetryOnCreate,
		utils.StateStopRetryOnCreate,
		func(ctx context.Context, spaceID api.SpaceId, nodePoolId api.NodePoolId, reqEditors ...api.RequestEditorFn) (*api.GetKubernetesNodePoolResponse, error) {
			return provider.Client.GetKubernetesNodePoolWithResponse(ctx, spaceID, parsedClusterID, nodePoolId, reqEditors...)
		},
		retryTimeout,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numspotNodePool, assert := read.(*api.KubernetesNodePool)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidKubernetesNodePoolAssertion, createdID)
	}

	return numspotNodePool, nil
}

// ReadKubernetesNodePool reads kubernetes node pool.
func ReadKubernetesNodePool(ctx context.Context, provider *client.NumspotSDK, clusterId, nodePoolId string) (*api.KubernetesNodePool, error) {
	parsedClusterID, err := uuid.Parse(clusterId)
	if err != nil {
		return nil, utils.WrapError(err, "failed to parse UUID")
	}
	parsedNodePoolID, err := uuid.Parse(nodePoolId)
	if err != nil {
		return nil, utils.WrapError(err, "failed to parse UUID")
	}

	res, err := provider.Client.GetKubernetesNodePoolWithResponse(ctx, provider.SpaceID, parsedClusterID, parsedNodePoolID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get kubernetes node pool")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get kubernetes node pool")
	}

	return res.JSON200, nil
}

// DeleteKubernetesNodePool deletes kubernetes node pool.
func DeleteKubernetesNodePool(ctx context.Context, provider *client.NumspotSDK, clusterId, nodePoolId string) (err error) {
	parsedClusterID, err := uuid.Parse(clusterId)
	if err != nil {
		return utils.WrapError(err, "failed to parse UUID")
	}
	parsedNodePoolID, err := uuid.Parse(nodePoolId)
	if err != nil {
		return utils.WrapError(err, "failed to parse UUID")
	}

	res, err := provider.Client.DeleteKubernetesNodePoolWithResponse(ctx, provider.SpaceID, parsedClusterID, parsedNodePoolID)
	if err != nil {
		return utils.WrapError(err, "failed to delete kubernetes node pool")
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return utils.WrapError(err, "failed to delete kubernetes node pool")
	}

	return nil
}
