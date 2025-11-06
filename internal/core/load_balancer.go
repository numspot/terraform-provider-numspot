package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateLoadBalancer(ctx context.Context, provider *client.NumSpotSDK, numSpotLoadBalancerCreate api.CreateLoadBalancerJSONRequestBody, numSpotLoadBalancerUpdate api.UpdateLoadBalancerJSONRequestBody, tags []api.ResourceTag, backendVM, backendIP []string) (numSpotVolume *api.LoadBalancer, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *api.CreateLoadBalancerResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotLoadBalancerCreate,
		numspotClient.CreateLoadBalancerWithResponse); err != nil {
		return nil, err
	}

	loadBalancerName := *retryCreate.JSON201.Name

	if len(backendIP) > 0 || len(backendVM) > 0 {
		if err = linkBackend(ctx, provider, loadBalancerName, backendIP, backendVM); err != nil {
			return nil, err
		}
	}

	if numSpotLoadBalancerUpdate.HealthCheck != nil || numSpotLoadBalancerUpdate.PublicIp != nil {
		if _, err = UpdateLoadBalancerAttributes(ctx, provider, loadBalancerName, numSpotLoadBalancerUpdate); err != nil {
			return nil, err
		}
	}

	if numSpotLoadBalancerUpdate.SecurityGroups != nil {
		if _, err = UpdateLoadBalancerSecurityGroup(ctx, provider, loadBalancerName, numSpotLoadBalancerUpdate); err != nil {
			return nil, err
		}
	}

	if len(numSpotLoadBalancerCreate.Listeners) > 0 {
		if err = createListeners(ctx, provider, loadBalancerName, numSpotLoadBalancerCreate.Listeners); err != nil {
			return nil, err
		}
	}

	if len(tags) > 0 {
		if err = createLoadBalancerTags(ctx, provider, loadBalancerName, tags); err != nil {
			return nil, err
		}
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

func DeleteLoadBalancer(ctx context.Context, provider *client.NumSpotSDK, loadBalancerID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, loadBalancerID, numspotClient.DeleteLoadBalancerWithResponse)
	if err != nil {
		return err
	}

	return nil
}

func ReadLoadBalancer(ctx context.Context, provider *client.NumSpotSDK, loadBalancerID string) (numSpotVolume *api.LoadBalancer, err error) {
	var numSpotReadLoadBalancer *api.ReadLoadBalancersByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadLoadBalancer, err = numspotClient.ReadLoadBalancersByIdWithResponse(ctx, provider.SpaceID, loadBalancerID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadLoadBalancer.Body, numSpotReadLoadBalancer.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadLoadBalancer.JSON200, err
}

func ReadLoadBalancers(ctx context.Context, provider *client.NumSpotSDK, loadBalancerParams api.ReadLoadBalancersParams) (numSpotLoadBalancers *[]api.LoadBalancer, err error) {
	var numSpotReadLoadBalancer *api.ReadLoadBalancersResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadLoadBalancer, err = numspotClient.ReadLoadBalancersWithResponse(ctx, provider.SpaceID, &loadBalancerParams)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadLoadBalancer.Body, numSpotReadLoadBalancer.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadLoadBalancer.JSON200.Items, err
}

func UpdateLoadBalancerAttributes(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, numSpotLoadBalancerUpdate api.UpdateLoadBalancerJSONRequestBody) (numSpotLoadBalancer *api.LoadBalancer, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var loadBalancerUpdateResponse *api.UpdateLoadBalancerResponse
	if loadBalancerUpdateResponse, err = numspotClient.UpdateLoadBalancerWithResponse(ctx, spaceID, loadBalancerName, api.UpdateLoadBalancerJSONRequestBody{
		HealthCheck: numSpotLoadBalancerUpdate.HealthCheck,
		PublicIp:    numSpotLoadBalancerUpdate.PublicIp,
	}); err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(loadBalancerUpdateResponse.Body, loadBalancerUpdateResponse.StatusCode()); err != nil {
		return nil, err
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

func UpdateLoadBalancerSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, numSpotLoadBalancerUpdate api.UpdateLoadBalancerJSONRequestBody) (numSpotLoadBalancer *api.LoadBalancer, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	securityGroups := make([]string, 0)
	if numSpotLoadBalancerUpdate.SecurityGroups != nil {
		securityGroups = *numSpotLoadBalancerUpdate.SecurityGroups
	}

	var loadBalancerUpdateResponse *api.UpdateLoadBalancerResponse
	if loadBalancerUpdateResponse, err = numspotClient.UpdateLoadBalancerWithResponse(ctx, spaceID, loadBalancerName, api.UpdateLoadBalancerJSONRequestBody{
		SecurityGroups: &securityGroups,
	}); err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(loadBalancerUpdateResponse.Body, loadBalancerUpdateResponse.StatusCode()); err != nil {
		return nil, err
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

func UpdateLoadBalancerTags(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, planTags []api.ResourceTag, stateTags []api.ResourceLoadBalancerTag) (numSpotLoadBalancer *api.LoadBalancer, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var loadBalancerDeleteTagsResponse *api.DeleteLoadBalancerTagsResponse
	if loadBalancerDeleteTagsResponse, err = numspotClient.DeleteLoadBalancerTagsWithResponse(ctx, spaceID, api.DeleteLoadBalancerTagsJSONRequestBody{
		Names: []string{loadBalancerName},
		Tags:  stateTags,
	}); err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(loadBalancerDeleteTagsResponse.Body, loadBalancerDeleteTagsResponse.StatusCode()); err != nil {
		return nil, err
	}

	var loadBalancerCreateTagsResponse *api.CreateLoadBalancerTagsResponse
	if loadBalancerCreateTagsResponse, err = numspotClient.CreateLoadBalancerTagsWithResponse(ctx, spaceID, api.CreateLoadBalancerTagsJSONRequestBody{
		Names: []string{loadBalancerName},
		Tags:  planTags,
	}); err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(loadBalancerCreateTagsResponse.Body, loadBalancerCreateTagsResponse.StatusCode()); err != nil {
		return nil, err
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

func UpdateLoadBalancerBackend(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, stateVMIDList, planVMIDList []string, stateIPList, planIPList []string) (numSpotLoadBalancer *api.LoadBalancer, err error) {
	if len(stateVMIDList) > 0 || len(stateIPList) > 0 {
		if err = unlinkBackend(ctx, provider, loadBalancerName, stateIPList, stateVMIDList); err != nil {
			return nil, err
		}
	}
	if len(planVMIDList) > 0 || len(planIPList) > 0 {
		if err = linkBackend(ctx, provider, loadBalancerName, planIPList, planVMIDList); err != nil {
			return nil, err
		}
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

func linkBackend(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, backendIP, backendVM []string) error {
	spaceID := provider.SpaceID
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	var bVM *[]string
	if backendVM != nil && len(backendVM) > 0 {
		bVM = &backendVM
	}

	var bIP *[]string
	if backendIP != nil && len(backendIP) > 0 {
		bIP = &backendIP
	}

	linkLB := api.LinkLoadBalancerBackendMachinesJSONRequestBody{
		BackendIps:   bIP,
		BackendVmIds: bVM,
	}

	linkLoadBalancerBackendResponse, err := numspotClient.LinkLoadBalancerBackendMachinesWithResponse(ctx, spaceID, loadBalancerName, linkLB)
	if err != nil {
		return err
	}
	if err = utils.ParseHTTPError(linkLoadBalancerBackendResponse.Body, linkLoadBalancerBackendResponse.StatusCode()); err != nil {
		return err
	}

	return nil
}

func unlinkBackend(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, backendIP, backendVM []string) error {
	spaceID := provider.SpaceID
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	bVM := make([]string, 0)
	if backendVM != nil {
		bVM = backendVM
	}

	bIP := make([]string, 0)
	if backendIP != nil {
		bIP = backendIP
	}

	unlinkLB := api.UnlinkLoadBalancerBackendMachinesJSONRequestBody{
		BackendIps:   &bIP,
		BackendVmIds: &bVM,
	}

	unlinkLoadBalancerBackendResponse, err := numspotClient.UnlinkLoadBalancerBackendMachinesWithResponse(ctx, spaceID, loadBalancerName, unlinkLB)
	if err != nil {
		return err
	}
	if err = utils.ParseHTTPError(unlinkLoadBalancerBackendResponse.Body, unlinkLoadBalancerBackendResponse.StatusCode()); err != nil {
		return err
	}

	return nil
}

func createLoadBalancerTags(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, tagList []api.ResourceTag) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	createTags := api.CreateLoadBalancerTagsRequest{
		Names: []string{loadBalancerName},
		Tags:  tagList,
	}

	var loadBalancerCreateTagsResponse *api.CreateLoadBalancerTagsResponse
	if loadBalancerCreateTagsResponse, err = numspotClient.CreateLoadBalancerTagsWithResponse(ctx, spaceID, createTags); err != nil {
		return err
	}
	if err = utils.ParseHTTPError(loadBalancerCreateTagsResponse.Body, loadBalancerCreateTagsResponse.StatusCode()); err != nil {
		return err
	}

	return nil
}

func createListeners(ctx context.Context, provider *client.NumSpotSDK, loadBalancerId string, listeners []api.ListenerForCreation) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	numSpotLoadBalancerListenersResponse, err := numspotClient.CreateLoadBalancerListenersWithResponse(ctx, provider.SpaceID, loadBalancerId,
		api.CreateLoadBalancerListenersJSONRequestBody{
			Listeners: listeners,
		})
	if err != nil {
		return err
	}

	if err = utils.ParseHTTPError(numSpotLoadBalancerListenersResponse.Body, numSpotLoadBalancerListenersResponse.StatusCode()); err != nil {
		return err
	}
	return err
}
