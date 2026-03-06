package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateSecurityGroupRule(ctx context.Context, provider *client.NumSpotSDK, id string, body api.CreateSecurityGroupRuleJSONRequestBody) (*api.SecurityGroup, error) {
	numSpotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numSpotClient.CreateSecurityGroupRuleWithResponse(ctx, provider.SpaceID, id, body)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON201, nil
}

func DeleteSecurityGroupRule(ctx context.Context, provider *client.NumSpotSDK, id string, body api.DeleteSecurityGroupRuleJSONRequestBody) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	res, err := numspotClient.DeleteSecurityGroupRuleWithResponse(ctx, provider.SpaceID, id, body)
	if err != nil {
		return err
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return err
	}

	return nil
}
