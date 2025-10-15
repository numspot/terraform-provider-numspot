package core

import (
	"context"
	"errors"
	"net/http"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
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

	switch res.StatusCode() {
	case http.StatusBadRequest:
		return nil, errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
	case http.StatusUnauthorized:
		return nil, errors.New(res.ApplicationproblemJSON401.Title + " : " + *res.ApplicationproblemJSON401.Detail)
	case http.StatusForbidden:
		return nil, errors.New(res.ApplicationproblemJSON403.Title + " : " + *res.ApplicationproblemJSON403.Detail)
	case http.StatusNotFound:
		return nil, errors.New(res.ApplicationproblemJSON404.Title + " : " + *res.ApplicationproblemJSON404.Detail)
	case http.StatusRequestURITooLong:
		return nil, errors.New(res.ApplicationproblemJSON414.Title + " : " + *res.ApplicationproblemJSON414.Detail)
	case http.StatusInternalServerError:
		return nil, errors.New(res.ApplicationproblemJSON500.Title)
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

	switch res.StatusCode() {
	case http.StatusBadRequest:
		return errors.New(res.ApplicationproblemJSON400.Title + " : " + *res.ApplicationproblemJSON400.Detail)
	case http.StatusUnauthorized:
		return errors.New(res.ApplicationproblemJSON401.Title + " : " + *res.ApplicationproblemJSON401.Detail)
	case http.StatusForbidden:
		return errors.New(res.ApplicationproblemJSON403.Title + " : " + *res.ApplicationproblemJSON403.Detail)
	case http.StatusNotFound:
		return errors.New(res.ApplicationproblemJSON404.Title + " : " + *res.ApplicationproblemJSON404.Detail)
	case http.StatusRequestURITooLong:
		return errors.New(res.ApplicationproblemJSON414.Title + " : " + *res.ApplicationproblemJSON414.Detail)
	case http.StatusInternalServerError:
		return errors.New(res.ApplicationproblemJSON500.Title)
	}

	return nil
}
