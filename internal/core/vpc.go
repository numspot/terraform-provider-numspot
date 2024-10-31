package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	vpcPendingStates = []string{pending, deleting}
	vpcTargetStates  = []string{available}
)

func CreateVPC(ctx context.Context, provider *client.NumSpotSDK, numSpotCreateVPC numspot.CreateVpcJSONRequestBody, dhcpOptionsSetID string, tags []numspot.ResourceTag) (*numspot.Vpc, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreateVpcResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotCreateVPC, numspotClient.CreateVpcWithResponse); err != nil {
		return nil, err
	}

	vpcID := *retryCreate.JSON201.Id

	if dhcpOptionsSetID != "" {
		var numSpotUpdateVPC *numspot.UpdateVpcResponse
		numSpotUpdateVPC, err = numspotClient.UpdateVpcWithResponse(ctx, spaceID, vpcID, numspot.UpdateVpcJSONRequestBody{DhcpOptionsSetId: dhcpOptionsSetID})
		if err != nil {
			return nil, err
		}
		if err = utils.ParseHTTPError(numSpotUpdateVPC.Body, numSpotUpdateVPC.StatusCode()); err != nil {
			return nil, err
		}
	}

	if len(tags) > 0 {
		if err = createTags(ctx, provider, vpcID, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadVPC(ctx, provider, createOp, vpcID)
}

func ReadVPC(ctx context.Context, provider *client.NumSpotSDK, vpcID string) (*numspot.Vpc, error) {
	var numSpotReadVPC *numspot.ReadVpcsByIdResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadVPC, err = numspotClient.ReadVpcsByIdWithResponse(ctx, provider.SpaceID, vpcID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadVPC.Body, numSpotReadVPC.StatusCode()); err != nil {
		return nil, err
	}
	return numSpotReadVPC.JSON200, nil
}

func RetryReadVPC(ctx context.Context, provider *client.NumSpotSDK, op string, vpcID string) (*numspot.Vpc, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, vpcID, provider.SpaceID, vpcPendingStates, vpcTargetStates, numspotClient.ReadVpcsByIdWithResponse)
	if err != nil {
		return nil, err
	}
	numSpotVPC, assert := read.(*numspot.Vpc)
	if !assert {
		return nil, fmt.Errorf("invalid vpc assertion %s: %s", vpcID, op)
	}
	return numSpotVPC, err
}

func UpdateVPCTags(ctx context.Context, provider *client.NumSpotSDK, volumeID string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.Vpc, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, volumeID); err != nil {
		return nil, err
	}
	return RetryReadVPC(ctx, provider, updateOp, volumeID)
}

func DeleteVPC(ctx context.Context, provider *client.NumSpotSDK, vpcID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, vpcID, numspotClient.DeleteVpcWithResponse)
}
