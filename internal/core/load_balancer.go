package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

const (
	kindVM = "kindVM"
	kindIP = "kindIP"
)

var (
	loadBalancerPendingStates = []string{}
	loadBalancerTargetStates  = []string{}
)

func CreateLoadBalancer(ctx context.Context, provider *client.NumSpotSDK, numSpotLoadBalancerCreate numspot.CreateLoadBalancerJSONRequestBody, tags []numspot.ResourceTag, healthCheck *numspot.HealthCheck, backendVM, backendIP []string) (numSpotVolume *numspot.LoadBalancer, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreateLoadBalancerResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotLoadBalancerCreate,
		numspotClient.CreateLoadBalancerWithResponse); err != nil {
		return nil, err
	}

	loadBalancerName := *retryCreate.JSON201.Name

	switch {
	case len(backendVM) > 0:
		if err = linkBackend(ctx, provider, loadBalancerName, kindVM, backendVM); err != nil {
			return nil, err
		}
	case len(backendIP) > 0:
		if err = linkBackend(ctx, provider, loadBalancerName, kindIP, backendIP); err != nil {
			return nil, err
		}
	}

	if healthCheck != nil {
		if err = attachHealthCheck(ctx, provider, loadBalancerName, healthCheck); err != nil {
			return nil, err
		}
	}

	if len(tags) > 0 {
		if err = CreateLoadBalancerTags(ctx, provider, loadBalancerName, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadLoadBalancer(ctx, provider, createOp, loadBalancerName)
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

func RetryReadLoadBalancer(ctx context.Context, provider *client.NumSpotSDK, op string, loadBalancerID string) (*numspot.LoadBalancer, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, loadBalancerID, provider.SpaceID, loadBalancerPendingStates, loadBalancerTargetStates, numspotClient.ReadLoadBalancersByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotLoadBalancer, assert := read.(*numspot.LoadBalancer)
	if !assert {
		return nil, fmt.Errorf("invalid load balancer assertion %s: %s", loadBalancerID, op)
	}
	return numSpotLoadBalancer, err
}

func ReadLoadBalancer(ctx context.Context, provider *client.NumSpotSDK, loadBalancerID string) (numSpotVolume *numspot.LoadBalancer, err error) {
	var numSpotReadLoadBalancer *numspot.ReadLoadBalancersByIdResponse
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

func linkBackend(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, backendKind string, backendToLink []string) error {
	spaceID := provider.SpaceID
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	linkLB := numspot.LinkLoadBalancerBackendMachinesJSONRequestBody{}

	switch backendKind {
	case kindVM:
		linkLB.BackendVmIds = &backendToLink
	case kindIP:
		linkLB.BackendIps = &backendToLink
	}

	if _, err = numspotClient.LinkLoadBalancerBackendMachinesWithResponse(ctx, spaceID, loadBalancerName, linkLB); err != nil {
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

	unlinkLB := numspot.UnlinkLoadBalancerBackendMachinesJSONRequestBody{
		BackendIps:   &backendIP,
		BackendVmIds: &backendVM,
	}

	if _, err = numspotClient.UnlinkLoadBalancerBackendMachinesWithResponse(ctx, spaceID, loadBalancerName, unlinkLB); err != nil {
		return err
	}
	return nil
}

func attachHealthCheck(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, healthCheck *numspot.HealthCheck) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	if _, err = numspotClient.UpdateLoadBalancerWithResponse(ctx, spaceID, loadBalancerName, numspot.UpdateLoadBalancerJSONRequestBody{HealthCheck: healthCheck}); err != nil {
		return err
	}

	return nil
}

func CreateLoadBalancerTags(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, tagList []numspot.ResourceTag) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	createTags := numspot.CreateLoadBalancerTagsRequest{
		Names: []string{loadBalancerName},
		Tags:  tagList,
	}

	var loadBalancerCreateTagsResponse *numspot.CreateLoadBalancerTagsResponse
	if loadBalancerCreateTagsResponse, err = numspotClient.CreateLoadBalancerTagsWithResponse(ctx, spaceID, createTags); err != nil {
		return err
	}
	if err = utils.ParseHTTPError(loadBalancerCreateTagsResponse.Body, loadBalancerCreateTagsResponse.StatusCode()); err != nil {
		return err
	}

	return nil
}

func DeleteLoadBalancerTags(ctx context.Context, provider *client.NumSpotSDK, loadBalancerName string, tagList []numspot.ResourceLoadBalancerTag) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	deleteTags := numspot.DeleteLoadBalancerTagsRequest{
		Names: []string{loadBalancerName},
		Tags:  tagList,
	}

	var loadBalancerDeleteTagsResponse *numspot.DeleteLoadBalancerTagsResponse
	if loadBalancerDeleteTagsResponse, err = numspotClient.DeleteLoadBalancerTagsWithResponse(ctx, spaceID, deleteTags); err != nil {
		return err
	}
	if err = utils.ParseHTTPError(loadBalancerDeleteTagsResponse.Body, loadBalancerDeleteTagsResponse.StatusCode()); err != nil {
		return err
	}

	return nil
}

//func DeleteLoadBalancerTags(
//	ctx context.Context,
//	spaceId numspot.SpaceId,
//	iaasClient *numspot.ClientWithResponses,
//	loadBalancerName string,
//	tagList types.List,
//	diags *diag.Diagnostics,
//) {
//	tfTags := make([]tags.TagsValue, 0, len(tagList.Elements()))
//	tagList.ElementsAs(ctx, &tfTags, false)
//
//	apiTags := make([]numspot.ResourceLoadBalancerTag, 0, len(tfTags))
//	for _, tfTag := range tfTags {
//		apiTags = append(apiTags, numspot.ResourceLoadBalancerTag{
//			Key: tfTag.Key.ValueStringPointer(),
//		})
//	}
//
//	_ = utils.ExecuteRequest(func() (*numspot.DeleteLoadBalancerTagsResponse, error) {
//		return iaasClient.DeleteLoadBalancerTagsWithResponse(ctx, spaceId, numspot.DeleteLoadBalancerTagsJSONRequestBody{
//			Names: []string{loadBalancerName},
//			Tags:  apiTags,
//		})
//	}, http.StatusNoContent, diags)
//}
