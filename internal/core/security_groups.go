package core

import (
	"context"
	"errors"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

// errEmptySecurityGroupResponse is returned when a security group read
// response unexpectedly has no JSON body.
var errEmptySecurityGroupResponse = errors.New("unexpected empty response body from ReadSecurityGroupsById")

// CreateSecurityGroup creates security group.
func CreateSecurityGroup(
	ctx context.Context,
	provider *client.NumSpotSDK,
	payload *api.CreateSecurityGroupJSONRequestBody,
	tags []api.ResourceTag,
) (numSpotSecurityGroup *api.SecurityGroup, err error) {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateSecurityGroupResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.SpaceID,
		*payload,
		numSpotClient.CreateSecurityGroupWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	securityGroupID := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, securityGroupID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	return ReadSecurityGroup(ctx, provider, securityGroupID)
}

// UpdateSecurityGroupTags updates security group tags.
func UpdateSecurityGroupTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	securityGroupID string,
	stateTags, planTags []api.ResourceTag,
) (*api.SecurityGroup, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, securityGroupID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	return ReadSecurityGroup(ctx, provider, securityGroupID)
}

// ReadSecurityGroup reads security group.
func ReadSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, id string) (*api.SecurityGroup, error) {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadSecurityGroup, err := numSpotClient.ReadSecurityGroupsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read security group")
	}
	if err = utils.ParseHTTPError(numSpotReadSecurityGroup.Body, numSpotReadSecurityGroup.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to parse HTTP error response")
	}
	if numSpotReadSecurityGroup.JSON200 == nil {
		return nil, errEmptySecurityGroupResponse
	}
	return numSpotReadSecurityGroup.JSON200, nil
}

// DeleteSecurityGroup deletes security group.
func DeleteSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, id string) error {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, numSpotClient.DeleteSecurityGroupWithResponse),
		"failed to delete resource",
	)
}

// ReadSecurityGroups reads security groups.
func ReadSecurityGroups(ctx context.Context, provider *client.NumSpotSDK, params *api.ReadSecurityGroupsParams) (*[]api.SecurityGroup, error) {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	res, err := numSpotClient.ReadSecurityGroupsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read security groups")
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to parse HTTP error response")
	}
	return res.JSON200.Items, nil
}
