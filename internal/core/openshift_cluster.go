package core

//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"time"
//
//	"github.com/google/uuid"
//	"terraform-provider-numspot/internal/client"
//	"terraform-provider-numspot/internal/sdk/api"
//)
//
//func CreateOpenshiftCluster(ctx context.Context, provider *client.NumSpotSDK, numSpotClusterCreate api.CreateClusterJSONRequestBody) (*api.OpenShiftCluster, error) {
//	spaceID := provider.SpaceID
//
//	numspotClient, err := provider.GetClient(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	res, err := numspotClient.CreateClusterWithResponse(ctx, spaceID, numSpotClusterCreate)
//	if err != nil {
//		return nil, err
//	}
//
//	if res.ApplicationproblemJSON400 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
//	}
//	if res.ApplicationproblemJSON500 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON500.Title + " : " + *res.ApplicationproblemJSON500.Detail)
//	}
//
//	return getOpenshiftClusterAfterCreation(ctx, provider, res.JSON201)
//}
//
//func getOpenshiftClusterAfterCreation(ctx context.Context, provider *client.NumSpotSDK, cluster *api.OpenShiftClusterCreated) (*api.OpenShiftCluster, error) {
//	const pollingInterval = 10 * time.Second
//	// maxRetries := 3 ?
//	loob := true
//	var message error
//
//	clusterIdString := cluster.Id.String()
//
//	for loob {
//		getClusterResponse, err := ReadOpenshiftCluster(ctx, provider, clusterIdString)
//		if err != nil {
//			loob = false
//			message = err
//		}
//
//		if getClusterResponse != nil {
//			state := *getClusterResponse.State
//
//			switch state {
//			case "N/A":
//				time.Sleep(pollingInterval)
//			case "CREATING":
//				time.Sleep(pollingInterval)
//			case "ACTIVE":
//				return getClusterResponse, nil
//			case "FAILED":
//				loob = false
//				message = errors.New("cluster creation failed")
//			default:
//				time.Sleep(pollingInterval)
//			}
//		}
//
//	}
//
//	return nil, message
//}
//
//func ReadClusters(ctx context.Context, provider *client.NumSpotSDK, clusters api.ListClustersParams) (*api.OpenShiftClusters, error) {
//	numspotClient, err := provider.GetClient(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	res, err := numspotClient.ListClustersWithResponse(ctx, provider.SpaceID, &clusters)
//	if err != nil {
//		return nil, err
//	}
//
//	if res.ApplicationproblemJSON400 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
//	}
//	if res.ApplicationproblemJSON500 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON500.Title + " : " + *res.ApplicationproblemJSON500.Detail)
//	}
//
//	return res.JSON200, nil
//}
//
//func ReadOpenshiftCluster(ctx context.Context, provider *client.NumSpotSDK, clusterId string) (*api.OpenShiftCluster, error) {
//	numspotClient, err := provider.GetClient(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	res, err := numspotClient.GetClusterWithResponse(ctx, provider.SpaceID, uuid.MustParse(clusterId))
//	if err != nil {
//		return nil, err
//	}
//
//	if res.ApplicationproblemJSON400 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
//	}
//	if res.ApplicationproblemJSON404 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON404.Title + " : " + *res.ApplicationproblemJSON404.Detail)
//	}
//	if res.ApplicationproblemJSON500 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON500.Title + " : " + *res.ApplicationproblemJSON500.Detail)
//	}
//
//	return res.JSON200, nil
//}
//
//func DeleteOpenshiftCluster(ctx context.Context, provider *client.NumSpotSDK, clusterId string) (numspotCluster *api.OpenShiftClusterDeleted, err error) {
//	spaceID := provider.SpaceID
//
//	numspotClient, err := provider.GetClient(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	res, err := numspotClient.DeleteClusterWithResponse(ctx, spaceID, uuid.MustParse(clusterId))
//	if err != nil {
//		return nil, err
//	}
//
//	if res.ApplicationproblemJSON400 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
//	}
//
//	if res.ApplicationproblemJSON404 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON404.Title + " : " + *res.ApplicationproblemJSON404.Detail)
//	}
//
//	if res.ApplicationproblemJSON500 != nil {
//		return nil, errors.New(res.ApplicationproblemJSON500.Title + " : " + *res.ApplicationproblemJSON500.Detail)
//	}
//
//	_, err = getOpenshiftOperationStateAfterDeletion(ctx, provider, res.JSON202)
//	if err != nil {
//		return numspotCluster, err
//	}
//
//	return res.JSON202, nil
//}
//
//func getOpenshiftOperationStateAfterDeletion(ctx context.Context, provider *client.NumSpotSDK, clusterDeletedRes *api.OpenShiftClusterDeleted) (*api.GetOperationResponse, error) {
//	const pollingInterval = 10 * time.Second
//	loob := true
//	var message error
//
//	for loob {
//		numspotClient, err := provider.GetClient(ctx)
//		if err != nil {
//			loob = false
//			message = fmt.Errorf("failed to get client: %w", err)
//		}
//
//		getOperationResponse, err := numspotClient.GetOperationWithResponse(ctx, provider.SpaceID, *clusterDeletedRes.Operation.Id)
//		if err != nil {
//			loob = false
//			message = fmt.Errorf("failed to get openshift operation: %w", err)
//		}
//
//		if getOperationResponse.ApplicationproblemJSON400 != nil {
//			loob = false
//			message = errors.New(getOperationResponse.ApplicationproblemJSON400.Title + " : " + *getOperationResponse.ApplicationproblemJSON400.Detail)
//		}
//		if getOperationResponse.ApplicationproblemJSON404 != nil {
//			loob = false
//			message = errors.New(getOperationResponse.ApplicationproblemJSON404.Title + " : " + *getOperationResponse.ApplicationproblemJSON404.Detail)
//		}
//		if getOperationResponse.ApplicationproblemJSON500 != nil {
//			loob = false
//			message = errors.New(getOperationResponse.ApplicationproblemJSON500.Title + " : " + *getOperationResponse.ApplicationproblemJSON500.Detail)
//		}
//
//		if getOperationResponse.HTTPResponse.StatusCode == 401 {
//			return getOperationResponse, nil
//		}
//
//		if getOperationResponse.JSON200 != nil {
//			state := *getOperationResponse.JSON200.Status
//
//			switch state {
//			case "PENDING":
//				time.Sleep(pollingInterval)
//			case "RUNNING":
//				time.Sleep(pollingInterval)
//			case "DONE":
//				return getOperationResponse, nil
//			case "FAILED":
//				loob = false
//				message = errors.New("cluster deletion failed")
//			default:
//				time.Sleep(pollingInterval)
//			}
//		}
//	}
//
//	return nil, message
//}
