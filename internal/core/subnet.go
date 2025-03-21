package core

import (
	"context"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	subnetPendingStates = []string{pending}
	subnetTargetStates  = []string{available}
)

func CreateSubnet(ctx context.Context, provider *client.NumSpotSDK, payload api.CreateSubnet, mapPublicIPOnLaunch bool, tags []api.ResourceTag) (*api.Subnet, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *api.CreateSubnetResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, payload, numspotClient.CreateSubnetWithResponse); err != nil {
		return nil, err
	}

	subnetID := *retryCreate.JSON201.Id

	if mapPublicIPOnLaunch {
		if _, err = UpdateSubnetAttributes(ctx, provider, subnetID, mapPublicIPOnLaunch); err != nil {
			return nil, err
		}
	}

	if len(tags) > 0 {
		if err = createTags(ctx, provider, subnetID, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadSubnet(ctx, provider, createOp, subnetID)
}

func UpdateSubnetAttributes(ctx context.Context, provider *client.NumSpotSDK, subnetID string, mapPublicIpOnLaunch bool) (*api.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var numSpotSubnet *api.UpdateSubnetResponse
	if numSpotSubnet, err = numspotClient.UpdateSubnetWithResponse(ctx, provider.SpaceID, subnetID,
		api.UpdateSubnetJSONRequestBody{
			MapPublicIpOnLaunch: mapPublicIpOnLaunch,
		},
	); err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotSubnet.Body, numSpotSubnet.StatusCode()); err != nil {
		return nil, err
	}

	return RetryReadSubnet(ctx, provider, updateOp, subnetID)
}

func UpdateSubnetTags(ctx context.Context, provider *client.NumSpotSDK, subnetID string, stateTags []api.ResourceTag, planTags []api.ResourceTag) (*api.Subnet, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, subnetID); err != nil {
		return nil, err
	}
	return RetryReadSubnet(ctx, provider, updateOp, subnetID)
}

func DeleteSubnet(ctx context.Context, provider *client.NumSpotSDK, subnetID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, subnetID, numspotClient.DeleteSubnetWithResponse)
}

func ReadSubnet(ctx context.Context, provider *client.NumSpotSDK, subnetID string) (*api.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	numSpotReadSubnet, err := numspotClient.ReadSubnetsByIdWithResponse(ctx, provider.SpaceID, subnetID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadSubnet.Body, numSpotReadSubnet.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadSubnet.JSON200, nil
}

func ReadSubnetsWithParams(ctx context.Context, provider *client.NumSpotSDK, params api.ReadSubnetsParams) (*[]api.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	numSpotReadSubnet, err := numspotClient.ReadSubnetsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadSubnet.Body, numSpotReadSubnet.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadSubnet.JSON200.Items, nil
}

func RetryReadSubnet(ctx context.Context, provider *client.NumSpotSDK, op, subnetID string) (*api.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, subnetID, provider.SpaceID, subnetPendingStates, subnetTargetStates,
		numspotClient.ReadSubnetsByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotSubnet, assert := read.(*api.Subnet)
	if !assert {
		return nil, fmt.Errorf("invalid client gateway assertion %s: %s", subnetID, op)
	}
	return numSpotSubnet, nil
}
