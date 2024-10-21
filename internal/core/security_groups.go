package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateSecurityGroup(
	ctx context.Context,
	provider *client.NumSpotSDK,
	payload numspot.CreateSecurityGroupJSONRequestBody,
	tags []numspot.ResourceTag,
	inboundRules,
	outboundRules *numspot.CreateSecurityGroupRuleJSONRequestBody,
) (*numspot.SecurityGroup, error) {
	client, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.SpaceID,
		payload,
		client.CreateSecurityGroupWithResponse)
	if err != nil {
		return nil, err
	}
	createdID := utils.GetPtrValue(res.JSON201.Id)
	if createdID == "" {
		return nil, errors.New("failed to get security group ID from response")
	}

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, createdID, tags); err != nil {
			return nil, fmt.Errorf("failed to update tags: %w", err)
		}
	}

	if err = updateAllRules(ctx, provider, createdID, inboundRules, outboundRules); err != nil {
		return nil, fmt.Errorf("failed to update rules: %w", err)
	}

	sg, err := ReadSecurityGroup(ctx, provider, createdID)
	if err != nil {
		return nil, err
	}

	return sg, nil
}

func UpdateSecurityGroupTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	id string,
	stateTags, planTags []numspot.ResourceTag,
) (*numspot.SecurityGroup, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, id); err != nil {
		return nil, err
	}
	return ReadSecurityGroup(ctx, provider, id)
}

func UpdateSecurityGroupRules(
	ctx context.Context,
	provider *client.NumSpotSDK,
	id string,
	inboundRules, outboundRules *numspot.CreateSecurityGroupRuleJSONRequestBody,
) (*numspot.SecurityGroup, error) {
	if err := updateAllRules(ctx, provider, id, inboundRules, outboundRules); err != nil {
		return nil, err
	}
	return ReadSecurityGroup(ctx, provider, id)
}

func ReadSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, id string) (*numspot.SecurityGroup, error) {
	client, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := client.ReadSecurityGroupsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, utils.HandleError(res.Body)
	}

	return res.JSON200, nil
}

func ReadSecurityGroups(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadSecurityGroupsParams) (*[]numspot.SecurityGroup, error) {
	client, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := client.ReadSecurityGroupsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200.Items, nil
}

func DeleteSecurityGroup(ctx context.Context, provider *client.NumSpotSDK, id string) error {
	client, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, client.DeleteSecurityGroupWithResponse)
}

func deleteRules(ctx context.Context, provider *client.NumSpotSDK, id string, existingRules *[]numspot.SecurityGroupRule, flow string) error {
	if existingRules == nil {
		return errors.New("failed to delete rules: rules cannot be nil")
	}

	client, err := provider.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete rules: %w", err)
	}

	rules := make([]numspot.SecurityGroupRule, 0, len(*existingRules))
	for _, e := range *existingRules {
		rules = append(rules, numspot.SecurityGroupRule{
			FromPortRange:         e.FromPortRange,
			IpProtocol:            e.IpProtocol,
			IpRanges:              e.IpRanges,
			SecurityGroupsMembers: e.SecurityGroupsMembers,
			ServiceIds:            e.ServiceIds,
			ToPortRange:           e.ToPortRange,
		})
	}

	payload := numspot.DeleteSecurityGroupRuleJSONRequestBody{
		Flow:  flow,
		Rules: &rules,
	}
	res, err := client.DeleteSecurityGroupRuleWithResponse(ctx, provider.SpaceID, id, payload)
	if err != nil {
		return fmt.Errorf("failed to delete rules: %w", err)
	}
	if res.StatusCode() != http.StatusNoContent {
		return utils.HandleError(res.Body)
	}

	return nil
}

func createRules(ctx context.Context, provider *client.NumSpotSDK, id string, rulesToCreate numspot.CreateSecurityGroupRuleJSONRequestBody) error {
	client, err := provider.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create rules: %w", err)
	}

	res, err := client.CreateSecurityGroupRuleWithResponse(ctx, provider.SpaceID, id, rulesToCreate)
	if err != nil {
		return fmt.Errorf("failed to create rules: %w", err)
	}
	if res.StatusCode() != http.StatusCreated {
		return fmt.Errorf("failed to create rules: %w", utils.HandleError(res.Body))
	}

	return nil
}

func updateAllRules(ctx context.Context, provider *client.NumSpotSDK, id string, inboundRules, outboundRules *numspot.CreateSecurityGroupRuleJSONRequestBody) error {
	// Read security group to retrieve the existing rules
	sg, err := ReadSecurityGroup(ctx, provider, id)
	if err != nil {
		return fmt.Errorf("failed to update rules: %w", err)
	}

	// Delete existing inbound rules
	if sg.InboundRules != nil && len(*sg.InboundRules) > 0 {
		if err = deleteRules(ctx, provider, id, sg.InboundRules, "Inbound"); err != nil {
			return fmt.Errorf("failed to update rules: %w", err)
		}
	}

	// Create wanted inbound rules
	if inboundRules != nil && len(*inboundRules.Rules) > 0 {
		if err = createRules(ctx, provider, id, *inboundRules); err != nil {
			return fmt.Errorf("failed to update rules: %w", err)
		}
	}

	// Delete existing Outbound rules
	if sg.OutboundRules != nil && len(*sg.OutboundRules) > 0 {
		if err = deleteRules(ctx, provider, id, sg.OutboundRules, "Outbound"); err != nil {
			return fmt.Errorf("failed to update rules: %w", err)
		}
	}
	// Create wanted Outbound rules
	if outboundRules != nil && len(*outboundRules.Rules) > 0 {
		if err = createRules(ctx, provider, id, *outboundRules); err != nil {
			return fmt.Errorf("failed to update rules: %w", err)
		}
	}

	return nil
}
