package core

import (
	"context"
	"errors"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	subnetPendingStates = []string{pending}
	subnetTargetStates  = []string{available}

	// errInvalidSubnetAssertion is returned when a subnet read result cannot be
	// asserted to the expected type. The message intentionally mirrors the
	// historical wording.
	errInvalidSubnetAssertion = errors.New("invalid client gateway assertion")
)

// CreateSubnet creates subnet.
func CreateSubnet(
	ctx context.Context,
	provider *client.NumSpotSDK,
	payload *api.CreateSubnet,
	mapPublicIPOnLaunch bool,
	tags []api.ResourceTag,
) (*api.Subnet, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateSubnetResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*payload,
		numspotClient.CreateSubnetWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	subnetID := *retryCreate.JSON201.Id

	if mapPublicIPOnLaunch {
		if _, err = UpdateSubnetAttributes(ctx, provider, subnetID, mapPublicIPOnLaunch); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	if len(tags) > 0 {
		if err = createTags(ctx, provider, subnetID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	return RetryReadSubnet(ctx, provider, createOp, subnetID)
}

// UpdateSubnetAttributes updates subnet attributes.
func UpdateSubnetAttributes(ctx context.Context, provider *client.NumSpotSDK, subnetID string, mapPublicIpOnLaunch bool) (*api.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var numSpotSubnet *api.UpdateSubnetResponse
	if numSpotSubnet, err = numspotClient.UpdateSubnetWithResponse(ctx, provider.SpaceID, subnetID,
		api.UpdateSubnetJSONRequestBody{
			MapPublicIpOnLaunch: mapPublicIpOnLaunch,
		},
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	if err = utils.ParseHTTPError(numSpotSubnet.Body, numSpotSubnet.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	return RetryReadSubnet(ctx, provider, updateOp, subnetID)
}

// UpdateSubnetTags updates subnet tags.
func UpdateSubnetTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	subnetID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.Subnet, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, subnetID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	return RetryReadSubnet(ctx, provider, updateOp, subnetID)
}

// DeleteSubnet deletes subnet.
func DeleteSubnet(ctx context.Context, provider *client.NumSpotSDK, subnetID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, subnetID, numspotClient.DeleteSubnetWithResponse),
		"failed to delete resource",
	)
}

// ReadSubnet reads subnet.
func ReadSubnet(ctx context.Context, provider *client.NumSpotSDK, subnetID string) (*api.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numSpotReadSubnet, err := numspotClient.ReadSubnetsByIdWithResponse(ctx, provider.SpaceID, subnetID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read subnets by id")
	}
	if err = utils.ParseHTTPError(numSpotReadSubnet.Body, numSpotReadSubnet.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read subnets by id")
	}

	return numSpotReadSubnet.JSON200, nil
}

// ReadSubnetsWithParams reads subnets with params.
func ReadSubnetsWithParams(ctx context.Context, provider *client.NumSpotSDK, params *api.ReadSubnetsParams) (*[]api.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	numSpotReadSubnet, err := numspotClient.ReadSubnetsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read subnets")
	}
	if err = utils.ParseHTTPError(numSpotReadSubnet.Body, numSpotReadSubnet.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read subnets")
	}

	return numSpotReadSubnet.JSON200.Items, nil
}

// RetryReadSubnet retries read subnet.
func RetryReadSubnet(ctx context.Context, provider *client.NumSpotSDK, op, subnetID string) (*api.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(ctx, subnetID, provider.SpaceID, subnetPendingStates, subnetTargetStates,
		numspotClient.ReadSubnetsByIdWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numSpotSubnet, assert := read.(*api.Subnet)
	if !assert {
		return nil, fmt.Errorf("%w %s: %s", errInvalidSubnetAssertion, subnetID, op)
	}
	return numSpotSubnet, nil
}
