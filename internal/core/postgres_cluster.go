package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
)

func CreatePostgresCluster(ctx context.Context, provider *client.NumSpotSDK, body api.PostgresClusterCreationRequestWithVolume) (*api.PostgresCluster, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	res, err := numspotClient.PostgreSQLCreateClusterWithBodyWithResponse(ctx, spaceID, "application/json", bytes.NewReader(bodyJson))
	if err != nil {
		return nil, err
	}

	if res.ApplicationproblemJSON400 != nil {
		return nil, errors.New("error creating cluster 400")
	}
	if res.ApplicationproblemJSON403 != nil {
		return nil, errors.New(res.ApplicationproblemJSON403.Title + " : " + *res.ApplicationproblemJSON403.Detail)
	}
	if res.ApplicationproblemJSON500 != nil {
		return nil, errors.New(res.ApplicationproblemJSON500.Title + " : " + *res.ApplicationproblemJSON500.Detail)
	}

	getClusterResponse, err := getClusterAfterCreation(ctx, provider, res.JSON201)
	if err != nil {
		return nil, err
	}

	return getClusterResponse, nil
}

func DeletePostgresCluster(ctx context.Context, provider *client.NumSpotSDK, clusterID api.PostgresClusterIdParameter, body api.PostgreSQLDeleteClusterJSONRequestBody) (*api.PostgresDeleteCluster202Response, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.PostgreSQLDeleteClusterWithResponse(ctx, spaceID, clusterID, body)
	if err != nil {
		return nil, err
	}

	if res.ApplicationproblemJSON400 != nil {
		return nil, errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
	}
	if res.ApplicationproblemJSON404 != nil {
		return nil, errors.New(res.ApplicationproblemJSON404.Title + " : " + *res.ApplicationproblemJSON404.Detail)
	}
	if res.ApplicationproblemJSON500 != nil {
		return nil, errors.New(res.ApplicationproblemJSON500.Title + " : " + *res.ApplicationproblemJSON500.Detail)
	}

	return res.JSON202, err
}

func ReadPostgresCluster(ctx context.Context, provider *client.NumSpotSDK, paramID api.PostgresClusterIdParameter) (*api.PostgresCluster, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.PostgreSQLGetClusterWithResponse(ctx, provider.SpaceID, paramID)
	if err != nil {
		return nil, err
	}

	if res.ApplicationproblemJSON400 != nil {
		return nil, errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
	}
	if res.ApplicationproblemJSON403 != nil {
		return nil, errors.New(res.ApplicationproblemJSON403.Title + " : " + *res.ApplicationproblemJSON403.Detail)
	}
	if res.ApplicationproblemJSON404 != nil {
		return nil, errors.New(res.ApplicationproblemJSON404.Title + " : " + *res.ApplicationproblemJSON404.Detail)
	}
	if res.ApplicationproblemJSON500 != nil {
		return nil, errors.New(res.ApplicationproblemJSON500.Title + " : " + *res.ApplicationproblemJSON500.Detail)
	}

	return res.JSON200, nil
}

func ReadPostgresClusters(ctx context.Context, provider *client.NumSpotSDK) (*api.PostgresListClusters200Response, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.PostgreSQLListClustersWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, err
	}

	if res.ApplicationproblemJSON400 != nil {
		return nil, errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
	}
	if res.ApplicationproblemJSON403 != nil {
		return nil, errors.New(res.ApplicationproblemJSON403.Title + " : " + *res.ApplicationproblemJSON403.Detail)
	}
	if res.ApplicationproblemJSON500 != nil {
		return nil, errors.New(res.ApplicationproblemJSON500.Title + " : " + *res.ApplicationproblemJSON500.Detail)
	}

	return res.JSON200, nil
}

func getClusterAfterCreation(ctx context.Context, provider *client.NumSpotSDK, cluster *api.PostgresClusterWithPassword) (*api.PostgresCluster, error) {
	const pollingInterval = 10 * time.Second
	loob := true
	var message error

	for loob {

		getClusterResponse, err := ReadPostgresCluster(ctx, provider, cluster.Id)
		if err != nil {
			loob = false
			message = err
		}

		if getClusterResponse != nil {
			state := getClusterResponse.Status

			switch state {
			case "CREATING":
				time.Sleep(pollingInterval)
			case "CONFIGURING":
				time.Sleep(pollingInterval)
			case "READY":
				return getClusterResponse, nil
			case "FAILED":
				loob = false
				message = errors.New("cluster creation failed")
			case "ERROR":
				loob = false
				message = errors.New("cluster creation error")
			default:
				time.Sleep(pollingInterval)
			}
		}
	}
	return nil, message
}
