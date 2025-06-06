package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, payload api.CreateSecurityGroupJSONRequestBody, tags []api.ResourceTag, inboundRules, outboundRules api.CreateSecurityGroupRuleJSONRequestBody) (numSpotSecurityGroup *api.SecurityGroup, err error) {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *api.CreateSecurityGroupResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, provider.SpaceID, payload, numSpotClient.CreateSecurityGroupWithResponse); err != nil {
		return nil, err
	}

	securityGroupID := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, securityGroupID, tags); err != nil {
			return nil, err
		}
	}

	defaultIPProtocol := "-1"
	defaultFromPortRange := -1
	defaultIPRanges := []string{"0.0.0.0/0"}
	defaultToPortRange := -1

	defaultRule := []api.SecurityGroupRule{
		{
			FromPortRange:         &defaultFromPortRange,
			IpProtocol:            &defaultIPProtocol,
			IpRanges:              &defaultIPRanges,
			SecurityGroupsMembers: nil,
			ServiceIds:            nil,
			ToPortRange:           &defaultToPortRange,
		},
	}

	if _, err = UpdateSecurityGroupRules(ctx, provider, securityGroupID,
		api.DeleteSecurityGroupRuleJSONRequestBody{},
		api.DeleteSecurityGroupRuleJSONRequestBody{
			Rules: &defaultRule,
			Flow:  "Outbound",
		},
		api.CreateSecurityGroupRuleJSONRequestBody{
			Rules: inboundRules.Rules,
			Flow:  inboundRules.Flow,
		},
		api.CreateSecurityGroupRuleJSONRequestBody{
			Rules: outboundRules.Rules,
			Flow:  outboundRules.Flow,
		},
	); err != nil {
		errDelete := DeleteSecurityGroup(ctx, provider, securityGroupID)
		if errDelete != nil {
			return nil, errDelete
		}
		return nil, err
	}

	return ReadSecurityGroup(ctx, provider, securityGroupID)
}

func UpdateSecurityGroupTags(ctx context.Context, provider *client.NumSpotSDK, securityGroupID string, stateTags, planTags []api.ResourceTag) (*api.SecurityGroup, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, securityGroupID); err != nil {
		return nil, err
	}
	return ReadSecurityGroup(ctx, provider, securityGroupID)
}

func UpdateSecurityGroupRules(ctx context.Context, provider *client.NumSpotSDK, securityGroupID string,
	stateInboundRules, stateOutboundRules api.DeleteSecurityGroupRuleJSONRequestBody,
	planInboundRules, planOutboundRules api.CreateSecurityGroupRuleJSONRequestBody,
) (numSpotSecurityGroup *api.SecurityGroup, err error) {
	if stateInboundRules.Rules != nil && len(*stateInboundRules.Rules) > 0 {
		if err = deleteRules(ctx, provider, securityGroupID, stateInboundRules); err != nil {
			return nil, err
		}
	}
	if stateOutboundRules.Rules != nil && len(*stateOutboundRules.Rules) > 0 {
		if err = deleteRules(ctx, provider, securityGroupID, stateOutboundRules); err != nil {
			return nil, err
		}
	}

	if planInboundRules.Rules != nil && len(*planInboundRules.Rules) > 0 {
		if err = createRules(ctx, provider, securityGroupID, planInboundRules); err != nil {
			return nil, err
		}
	}
	if planOutboundRules.Rules != nil && len(*planOutboundRules.Rules) > 0 {
		if err = createRules(ctx, provider, securityGroupID, planOutboundRules); err != nil {
			return nil, err
		}
	}

	return ReadSecurityGroup(ctx, provider, securityGroupID)
}

func ReadSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, id string) (*api.SecurityGroup, error) {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadSecurityGroup, err := numSpotClient.ReadSecurityGroupsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadSecurityGroup.Body, numSpotReadSecurityGroup.StatusCode()); err != nil {
		return nil, err
	}
	return numSpotReadSecurityGroup.JSON200, nil
}

func DeleteSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, id string) error {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, numSpotClient.DeleteSecurityGroupWithResponse)
}

func deleteRules(ctx context.Context, provider *client.NumSpotSDK, id string, rulesToDelete api.DeleteSecurityGroupRuleJSONRequestBody) error {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	numSpotDeleteSecurityGroupRule, err := numSpotClient.DeleteSecurityGroupRuleWithResponse(ctx, provider.SpaceID, id, rulesToDelete)
	if err != nil {
		return err
	}
	if err = utils.ParseHTTPError(numSpotDeleteSecurityGroupRule.Body, numSpotDeleteSecurityGroupRule.StatusCode()); err != nil {
		return err
	}

	return nil
}

func createRules(ctx context.Context, provider *client.NumSpotSDK, id string, rulesToCreate api.CreateSecurityGroupRuleJSONRequestBody) error {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	numSpotCreateSecurityGroupRule, err := numSpotClient.CreateSecurityGroupRuleWithResponse(ctx, provider.SpaceID, id, rulesToCreate)
	if err != nil {
		return err
	}
	if err = utils.ParseHTTPError(numSpotCreateSecurityGroupRule.Body, numSpotCreateSecurityGroupRule.StatusCode()); err != nil {
		return err
	}
	return nil
}

func ReadSecurityGroups(ctx context.Context, provider *client.NumSpotSDK, params api.ReadSecurityGroupsParams) (*[]api.SecurityGroup, error) {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	res, err := numSpotClient.ReadSecurityGroupsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}
	return res.JSON200.Items, nil
}
