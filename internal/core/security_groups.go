package core

import (
	"context"
	"errors"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// errEmptySecurityGroupResponse is returned when a security group read
// response unexpectedly has no JSON body.
var errEmptySecurityGroupResponse = errors.New("unexpected empty response body from ReadSecurityGroupsById")

// CreateSecurityGroup creates security group.
func CreateSecurityGroup(
	ctx context.Context,
	provider *client.NumspotSDK,
	payload *api.CreateSecurityGroupJSONRequestBody,
	tags []api.ResourceTag,
) (numspotSecurityGroup *api.SecurityGroup, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateSecurityGroupResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.SpaceID,
		*payload,
		numspotClient.CreateSecurityGroupWithResponse,
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
	provider *client.NumspotSDK,
	securityGroupID string,
	stateTags, planTags []api.ResourceTag,
) (*api.SecurityGroup, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, securityGroupID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	return ReadSecurityGroup(ctx, provider, securityGroupID)
}

// ReadSecurityGroup reads security group.
func ReadSecurityGroup(ctx context.Context, provider *client.NumspotSDK, id string) (*api.SecurityGroup, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadSecurityGroup, err := numspotClient.ReadSecurityGroupsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read security group")
	}
	if err = utils.ParseHTTPError(numspotReadSecurityGroup.Body, numspotReadSecurityGroup.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to parse HTTP error response")
	}
	if numspotReadSecurityGroup.JSON200 == nil {
		return nil, errEmptySecurityGroupResponse
	}
	return numspotReadSecurityGroup.JSON200, nil
}

// DeleteSecurityGroup deletes security group.
func DeleteSecurityGroup(ctx context.Context, provider *client.NumspotSDK, id string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, numspotClient.DeleteSecurityGroupWithResponse),
		"failed to delete resource",
	)
}

// ReadSecurityGroups reads security groups.
func ReadSecurityGroups(ctx context.Context, provider *client.NumspotSDK, params *api.ReadSecurityGroupsParams) (*[]api.SecurityGroup, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	res, err := numspotClient.ReadSecurityGroupsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read security groups")
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to parse HTTP error response")
	}
	return res.JSON200.Items, nil
}
