package core

import (
	"context"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// CreateLoadBalancer creates load balancer.
func CreateLoadBalancer(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotLoadBalancerCreate *api.CreateLoadBalancerJSONRequestBody,
	numspotLoadBalancerUpdate *api.UpdateLoadBalancerJSONRequestBody,
	backendVM, backendIP []string,
	tags []api.ResourceTag,
) (*api.LoadBalancer, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	retryCreate, err := utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, *numspotLoadBalancerCreate,
		numspotClient.CreateLoadBalancerWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	loadBalancerName := *retryCreate.JSON201.Name

	if err := linkBackendIfNeeded(ctx, provider, loadBalancerName, backendIP, backendVM); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err := updateLoadBalancerAttributesIfNeeded(ctx, provider, loadBalancerName, numspotLoadBalancerUpdate); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err := updateLoadBalancerSecurityGroupIfNeeded(ctx, provider, loadBalancerName, numspotLoadBalancerUpdate); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err := createLoadBalancerListenersIfNeeded(ctx, provider, loadBalancerName, numspotLoadBalancerCreate.Listeners); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err := createLoadBalancerTagsIfNeeded(ctx, provider, loadBalancerName, tags); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

func linkBackendIfNeeded(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	backendIP, backendVM []string,
) error {
	if len(backendIP) == 0 && len(backendVM) == 0 {
		return nil
	}

	return linkBackend(ctx, provider, loadBalancerName, backendIP, backendVM)
}

func updateLoadBalancerAttributesIfNeeded(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	update *api.UpdateLoadBalancerJSONRequestBody,
) error {
	if update.HealthCheck == nil && update.PublicIp == nil {
		return nil
	}

	_, err := UpdateLoadBalancerAttributes(ctx, provider, loadBalancerName, update)
	return err
}

func updateLoadBalancerSecurityGroupIfNeeded(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	update *api.UpdateLoadBalancerJSONRequestBody,
) error {
	if update.SecurityGroups == nil {
		return nil
	}

	_, err := UpdateLoadBalancerSecurityGroup(ctx, provider, loadBalancerName, update)
	return err
}

func createLoadBalancerListenersIfNeeded(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	listeners []api.ListenerForCreation,
) error {
	if len(listeners) == 0 {
		return nil
	}

	return createListeners(ctx, provider, loadBalancerName, listeners)
}

func createLoadBalancerTagsIfNeeded(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	tags []api.ResourceTag,
) error {
	if len(tags) == 0 {
		return nil
	}

	return createLoadBalancerTags(ctx, provider, loadBalancerName, tags)
}

// DeleteLoadBalancer deletes load balancer.
func DeleteLoadBalancer(ctx context.Context, provider *client.NumspotSDK, loadBalancerID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, loadBalancerID, numspotClient.DeleteLoadBalancerWithResponse)
	if err != nil {
		return utils.WrapError(err, "failed to delete resource")
	}

	return nil
}

// ReadLoadBalancer reads load balancer.
func ReadLoadBalancer(ctx context.Context, provider *client.NumspotSDK, loadBalancerID string) (numspotVolume *api.LoadBalancer, err error) {
	var numspotReadLoadBalancer *api.ReadLoadBalancersByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadLoadBalancer, err = numspotClient.ReadLoadBalancersByIdWithResponse(ctx, provider.SpaceID, loadBalancerID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read load balancers by id")
	}
	if err = utils.ParseHTTPError(numspotReadLoadBalancer.Body, numspotReadLoadBalancer.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read load balancers by id")
	}

	return numspotReadLoadBalancer.JSON200, utils.WrapError(err, "failed to read load balancers by id")
}

// ReadLoadBalancers reads load balancers.
func ReadLoadBalancers(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerParams *api.ReadLoadBalancersParams,
) (numspotLoadBalancers *[]api.LoadBalancer, err error) {
	var numspotReadLoadBalancer *api.ReadLoadBalancersResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadLoadBalancer, err = numspotClient.ReadLoadBalancersWithResponse(ctx, provider.SpaceID, loadBalancerParams)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read load balancers")
	}
	if err = utils.ParseHTTPError(numspotReadLoadBalancer.Body, numspotReadLoadBalancer.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read load balancers")
	}

	return numspotReadLoadBalancer.JSON200.Items, utils.WrapError(err, "failed to read load balancers")
}

// UpdateLoadBalancerAttributes updates load balancer attributes.
func UpdateLoadBalancerAttributes(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	numspotLoadBalancerUpdate *api.UpdateLoadBalancerJSONRequestBody,
) (numspotLoadBalancer *api.LoadBalancer, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var loadBalancerUpdateResponse *api.UpdateLoadBalancerResponse
	if loadBalancerUpdateResponse, err = numspotClient.UpdateLoadBalancerWithResponse(
		ctx,
		spaceID,
		loadBalancerName,
		api.UpdateLoadBalancerJSONRequestBody{
			HealthCheck: numspotLoadBalancerUpdate.HealthCheck,
			PublicIp:    numspotLoadBalancerUpdate.PublicIp,
		},
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err = utils.ParseHTTPError(loadBalancerUpdateResponse.Body, loadBalancerUpdateResponse.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

// UpdateLoadBalancerSecurityGroup updates load balancer security group.
func UpdateLoadBalancerSecurityGroup(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	numspotLoadBalancerUpdate *api.UpdateLoadBalancerJSONRequestBody,
) (numspotLoadBalancer *api.LoadBalancer, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	securityGroups := make([]string, 0)
	if numspotLoadBalancerUpdate.SecurityGroups != nil {
		securityGroups = *numspotLoadBalancerUpdate.SecurityGroups
	}

	var loadBalancerUpdateResponse *api.UpdateLoadBalancerResponse
	if loadBalancerUpdateResponse, err = numspotClient.UpdateLoadBalancerWithResponse(
		ctx,
		spaceID,
		loadBalancerName,
		api.UpdateLoadBalancerJSONRequestBody{
			SecurityGroups: &securityGroups,
		},
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if err = utils.ParseHTTPError(loadBalancerUpdateResponse.Body, loadBalancerUpdateResponse.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

// UpdateLoadBalancerTags updates load balancer tags.
func UpdateLoadBalancerTags(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (numspotLoadBalancer *api.LoadBalancer, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	tagsToDelete := make([]api.ResourceLoadBalancerTag, len(stateTags))
	for idx := range stateTags {
		key := stateTags[idx].Key
		tagsToDelete[idx] = api.ResourceLoadBalancerTag{Key: &key}
	}

	var loadBalancerDeleteTagsResponse *api.DeleteLoadBalancerTagsResponse
	if loadBalancerDeleteTagsResponse, err = numspotClient.DeleteLoadBalancerTagsWithResponse(ctx, spaceID, api.DeleteLoadBalancerTagsJSONRequestBody{
		Names: []string{loadBalancerName},
		Tags:  tagsToDelete,
	}); err != nil {
		return nil, utils.WrapError(err, "failed to delete load balancer tags")
	}

	if err = utils.ParseHTTPError(loadBalancerDeleteTagsResponse.Body, loadBalancerDeleteTagsResponse.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to delete load balancer tags")
	}

	var loadBalancerCreateTagsResponse *api.CreateLoadBalancerTagsResponse
	if loadBalancerCreateTagsResponse, err = numspotClient.CreateLoadBalancerTagsWithResponse(ctx, spaceID, api.CreateLoadBalancerTagsJSONRequestBody{
		Names: []string{loadBalancerName},
		Tags:  planTags,
	}); err != nil {
		return nil, utils.WrapError(err, "failed to create load balancer tags")
	}

	if err = utils.ParseHTTPError(loadBalancerCreateTagsResponse.Body, loadBalancerCreateTagsResponse.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to create load balancer tags")
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

// UpdateLoadBalancerBackend updates load balancer backend.
func UpdateLoadBalancerBackend(
	ctx context.Context,
	provider *client.NumspotSDK,
	loadBalancerName string,
	stateVMIDList, planVMIDList []string,
	stateIPList, planIPList []string,
) (numspotLoadBalancer *api.LoadBalancer, err error) {
	if len(stateVMIDList) > 0 || len(stateIPList) > 0 {
		if err = unlinkBackend(ctx, provider, loadBalancerName, stateIPList, stateVMIDList); err != nil {
			return nil, utils.WrapError(err, "failed to create load balancer tags")
		}
	}
	if len(planVMIDList) > 0 || len(planIPList) > 0 {
		if err = linkBackend(ctx, provider, loadBalancerName, planIPList, planVMIDList); err != nil {
			return nil, utils.WrapError(err, "failed to create load balancer tags")
		}
	}

	return ReadLoadBalancer(ctx, provider, loadBalancerName)
}

func linkBackend(ctx context.Context, provider *client.NumspotSDK, loadBalancerName string, backendIP, backendVM []string) error {
	spaceID := provider.SpaceID
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	var bVM *[]string
	if len(backendVM) > 0 {
		bVM = &backendVM
	}

	var bIP *[]string
	if len(backendIP) > 0 {
		bIP = &backendIP
	}

	linkLB := api.LinkLoadBalancerBackendMachinesJSONRequestBody{
		BackendIps:   bIP,
		BackendVmIds: bVM,
	}

	linkLoadBalancerBackendResponse, err := numspotClient.LinkLoadBalancerBackendMachinesWithResponse(ctx, spaceID, loadBalancerName, linkLB)
	if err != nil {
		return utils.WrapError(err, "failed to link load balancer backend machines")
	}
	if err = utils.ParseHTTPError(linkLoadBalancerBackendResponse.Body, linkLoadBalancerBackendResponse.StatusCode()); err != nil {
		return utils.WrapError(err, "failed to link load balancer backend machines")
	}

	return nil
}

func unlinkBackend(ctx context.Context, provider *client.NumspotSDK, loadBalancerName string, backendIP, backendVM []string) error {
	spaceID := provider.SpaceID
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
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
		return utils.WrapError(err, "failed to unlink load balancer backend machines")
	}
	if err = utils.ParseHTTPError(unlinkLoadBalancerBackendResponse.Body, unlinkLoadBalancerBackendResponse.StatusCode()); err != nil {
		return utils.WrapError(err, "failed to unlink load balancer backend machines")
	}

	return nil
}

func createLoadBalancerTags(ctx context.Context, provider *client.NumspotSDK, loadBalancerName string, tagList []api.ResourceTag) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	createTags := api.CreateLoadBalancerTagsRequest{
		Names: []string{loadBalancerName},
		Tags:  tagList,
	}

	var loadBalancerCreateTagsResponse *api.CreateLoadBalancerTagsResponse
	if loadBalancerCreateTagsResponse, err = numspotClient.CreateLoadBalancerTagsWithResponse(ctx, spaceID, createTags); err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	if err = utils.ParseHTTPError(loadBalancerCreateTagsResponse.Body, loadBalancerCreateTagsResponse.StatusCode()); err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	return nil
}

func createListeners(ctx context.Context, provider *client.NumspotSDK, loadBalancerId string, listeners []api.ListenerForCreation) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	numspotLoadBalancerListenersResponse, err := numspotClient.CreateLoadBalancerListenersWithResponse(ctx, provider.SpaceID, loadBalancerId,
		api.CreateLoadBalancerListenersJSONRequestBody{
			Listeners: listeners,
		})
	if err != nil {
		return utils.WrapError(err, "failed to create load balancer listeners")
	}

	if err = utils.ParseHTTPError(numspotLoadBalancerListenersResponse.Body, numspotLoadBalancerListenersResponse.StatusCode()); err != nil {
		return utils.WrapError(err, "failed to create load balancer listeners")
	}
	return utils.WrapError(err, "failed to create load balancer listeners")
}
