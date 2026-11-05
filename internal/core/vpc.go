package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
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
	provider *client.NumspotSDK,
	numspotCreateVPC *api.CreateVpcJSONRequestBody,
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
		*numspotCreateVPC,
		numspotClient.CreateVpcWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	vpcID := *retryCreate.JSON201.Id

	if dhcpOptionsSetID != "" {
		var numspotUpdateVPC *api.UpdateVpcResponse
		numspotUpdateVPC, err = numspotClient.UpdateVpcWithResponse(
			ctx,
			spaceID,
			vpcID,
			api.UpdateVpcJSONRequestBody{DhcpOptionsSetId: dhcpOptionsSetID},
		)
		if err != nil {
			return nil, utils.WrapError(err, "failed to update vpc")
		}
		if err = utils.ParseHTTPError(numspotUpdateVPC.Body, numspotUpdateVPC.StatusCode()); err != nil {
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
func ReadVPC(ctx context.Context, provider *client.NumspotSDK, vpcID string) (*api.Vpc, error) {
	var numspotReadVPC *api.ReadVpcsByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadVPC, err = numspotClient.ReadVpcsByIdWithResponse(ctx, provider.SpaceID, vpcID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vpcs by id")
	}
	if err = utils.ParseHTTPError(numspotReadVPC.Body, numspotReadVPC.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vpcs by id")
	}
	return numspotReadVPC.JSON200, nil
}

// ReadVPCsWithParams reads VPCs with params.
func ReadVPCsWithParams(ctx context.Context, provider *client.NumspotSDK, params *api.ReadVpcsParams) (*[]api.Vpc, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadVPC, err := numspotClient.ReadVpcsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vpcs")
	}
	if err = utils.ParseHTTPError(numspotReadVPC.Body, numspotReadVPC.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vpcs")
	}
	return numspotReadVPC.JSON200.Items, nil
}

// RetryReadVPC retries read VPC.
func RetryReadVPC(ctx context.Context, provider *client.NumspotSDK, _, vpcID string) (*api.Vpc, error) {
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
	numspotVPC, assert := read.(*api.Vpc)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidVPCAssertion, vpcID)
	}
	return numspotVPC, utils.WrapError(err, "failed to wait for resource state")
}

// UpdateVPCTags updates VPC tags.
func UpdateVPCTags(
	ctx context.Context,
	provider *client.NumspotSDK,
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
func DeleteVPC(ctx context.Context, provider *client.NumspotSDK, vpcID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, vpcID, numspotClient.DeleteVpcWithResponse),
		"failed to delete resource",
	)
}
