package core

import (
	"context"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, payload numspot.CreateSecurityGroupJSONRequestBody, tags []numspot.ResourceTag, inboundRules, outboundRules *numspot.CreateSecurityGroupRuleJSONRequestBody) (numSpotSecurityGroup *numspot.SecurityGroup, err error) {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreateSecurityGroupResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, provider.SpaceID, payload, numSpotClient.CreateSecurityGroupWithResponse); err != nil {
		return nil, err
	}

	securityGroupID := *retryCreate.JSON201.Id

	sg, err := ReadSecurityGroup(ctx, provider, securityGroupID)
	if err != nil {
		return nil, err
	}

	if len(tags) > 0 {
		if err = createTags(ctx, provider, securityGroupID, tags); err != nil {
			return nil, err
		}
	}

	defaultRuleToDelete := numspot.DeleteSecurityGroupRuleJSONRequestBody{
		Flow:  "Outbound",
		Rules: sg.OutboundRules,
	}

	_, err = UpdateSecurityGroupRules(ctx, provider, securityGroupID, *inboundRules, *outboundRules, nil, &defaultRuleToDelete)
	if err != nil {
		return nil, err
	}

	return ReadSecurityGroup(ctx, provider, securityGroupID)
}

func UpdateSecurityGroupTags(ctx context.Context, provider *client.NumSpotSDK, securityGroupID string, stateTags, planTags []numspot.ResourceTag) (*numspot.SecurityGroup, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, securityGroupID); err != nil {
		return nil, err
	}
	return ReadSecurityGroup(ctx, provider, securityGroupID)
}

func UpdateSecurityGroupRules(ctx context.Context, provider *client.NumSpotSDK, securityGroupID string,
	planInboundRules, planOutboundRules numspot.CreateSecurityGroupRuleJSONRequestBody,
	stateInboundRules, stateOutboundRules *numspot.DeleteSecurityGroupRuleJSONRequestBody,
) (numSpotSecurityGroup *numspot.SecurityGroup, err error) {
	if stateInboundRules != nil && stateInboundRules.Rules != nil && len(*stateInboundRules.Rules) > 0 {
		if err = deleteRules(ctx, provider, securityGroupID, *stateInboundRules); err != nil {
			return nil, err
		}
	}

	if planInboundRules.Rules != nil && len(*planInboundRules.Rules) > 0 {
		if err = createRules(ctx, provider, securityGroupID, planInboundRules); err != nil {
			return nil, err
		}
	}

	if stateOutboundRules != nil && stateOutboundRules.Rules != nil && len(*stateOutboundRules.Rules) > 0 {
		if err = deleteRules(ctx, provider, securityGroupID, *stateOutboundRules); err != nil {
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

func ReadSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, id string) (*numspot.SecurityGroup, error) {
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

func ReadSecurityGroups(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadSecurityGroupsParams) (*[]numspot.SecurityGroup, error) {
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

func DeleteSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, id string) error {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, numSpotClient.DeleteSecurityGroupWithResponse)
}

func deleteRules(ctx context.Context, provider *client.NumSpotSDK, id string, rulesToDelete numspot.DeleteSecurityGroupRuleJSONRequestBody) error {
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

func createRules(ctx context.Context, provider *client.NumSpotSDK, id string, rulesToCreate numspot.CreateSecurityGroupRuleJSONRequestBody) error {
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
