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
	vpcPendingStates = []string{pending, deleting}
	vpcTargetStates  = []string{available}

	// errInvalidVPCAssertion is returned when a VPC read result cannot be
	// asserted to the expected type.
	errInvalidVPCAssertion = errors.New("invalid vpc assertion")
)

// CreateVPC creates VPC.
func CreateVPC(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotCreateVPC *api.CreateVpcJSONRequestBody,
	dhcpOptionsSetID string,
	tags []api.ResourceTag,
) (*api.Vpc, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateVpcResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*numSpotCreateVPC,
		numspotClient.CreateVpcWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	vpcID := *retryCreate.JSON201.Id

	if dhcpOptionsSetID != "" {
		var numSpotUpdateVPC *api.UpdateVpcResponse
		numSpotUpdateVPC, err = numspotClient.UpdateVpcWithResponse(
			ctx,
			spaceID,
			vpcID,
			api.UpdateVpcJSONRequestBody{DhcpOptionsSetId: dhcpOptionsSetID},
		)
		if err != nil {
			return nil, utils.WrapError(err, "failed to update vpc")
		}
		if err = utils.ParseHTTPError(numSpotUpdateVPC.Body, numSpotUpdateVPC.StatusCode()); err != nil {
			return nil, utils.WrapError(err, "failed to update vpc")
		}
	}

	if len(tags) > 0 {
		if err = createTags(ctx, provider, vpcID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to update vpc")
		}
	}

	return RetryReadVPC(ctx, provider, createOp, vpcID)
}

// ReadVPC reads VPC.
func ReadVPC(ctx context.Context, provider *client.NumSpotSDK, vpcID string) (*api.Vpc, error) {
	var numSpotReadVPC *api.ReadVpcsByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadVPC, err = numspotClient.ReadVpcsByIdWithResponse(ctx, provider.SpaceID, vpcID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vpcs by id")
	}
	if err = utils.ParseHTTPError(numSpotReadVPC.Body, numSpotReadVPC.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vpcs by id")
	}
	return numSpotReadVPC.JSON200, nil
}

// ReadVPCsWithParams reads VPCs with params.
func ReadVPCsWithParams(ctx context.Context, provider *client.NumSpotSDK, params *api.ReadVpcsParams) (*[]api.Vpc, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadVPC, err := numspotClient.ReadVpcsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vpcs")
	}
	if err = utils.ParseHTTPError(numSpotReadVPC.Body, numSpotReadVPC.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vpcs")
	}
	return numSpotReadVPC.JSON200.Items, nil
}

// RetryReadVPC retries read VPC.
func RetryReadVPC(ctx context.Context, provider *client.NumSpotSDK, _, vpcID string) (*api.Vpc, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		vpcID,
		provider.SpaceID,
		vpcPendingStates,
		vpcTargetStates,
		numspotClient.ReadVpcsByIdWithResponse,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}
	numSpotVPC, assert := read.(*api.Vpc)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidVPCAssertion, vpcID)
	}
	return numSpotVPC, utils.WrapError(err, "failed to wait for resource state")
}

// UpdateVPCTags updates VPC tags.
func UpdateVPCTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	volumeID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.Vpc, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, volumeID); err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}
	return RetryReadVPC(ctx, provider, updateOp, volumeID)
}

// DeleteVPC deletes VPC.
func DeleteVPC(ctx context.Context, provider *client.NumSpotSDK, vpcID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, vpcID, numspotClient.DeleteVpcWithResponse),
		"failed to delete resource",
	)
}
