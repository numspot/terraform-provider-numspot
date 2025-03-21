package core

import (
	"context"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateDHCPOptions(ctx context.Context, provider *client.NumSpotSDK, numSpotDHCPOptionsCreate api.CreateDhcpOptionsJSONRequestBody, tags []api.ResourceTag) (numSpotDHCPOptions *api.DhcpOptionsSet, err error) {
	spaceID := provider.SpaceID

	var retryCreate *api.CreateDhcpOptionsResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotDHCPOptionsCreate, numspotClient.CreateDhcpOptionsWithResponse); err != nil {
		return nil, err
	}

	dhcpOptionsID := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, dhcpOptionsID, tags); err != nil {
			return nil, err
		}
	}

	return ReadDHCPOption(ctx, provider, dhcpOptionsID)
}

func UpdateDHCPOptionsTags(ctx context.Context, provider *client.NumSpotSDK, dhcpOptionsID string, stateTags []api.ResourceTag, planTags []api.ResourceTag) (*api.DhcpOptionsSet, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, dhcpOptionsID); err != nil {
		return nil, err
	}
	return ReadDHCPOption(ctx, provider, dhcpOptionsID)
}

func DeleteDHCPOptions(ctx context.Context, provider *client.NumSpotSDK, dhcpOptionsID string) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	if err := utils.RetryDeleteUntilResourceAvailable(ctx, spaceID, dhcpOptionsID,
		numspotClient.DeleteDhcpOptionsWithResponse); err != nil {
		return err
	}

	return nil
}

func ReadDHCPOptions(ctx context.Context, provider *client.NumSpotSDK, dhcpOptions api.ReadDhcpOptionsParams) (*[]api.DhcpOptionsSet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := numspotClient.ReadDhcpOptionsWithResponse(ctx, provider.SpaceID, &dhcpOptions)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200.Items, nil
}

func ReadDHCPOption(ctx context.Context, provider *client.NumSpotSDK, dhcpOptionID string) (*api.DhcpOptionsSet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := numspotClient.ReadDhcpOptionsByIdWithResponse(ctx, provider.SpaceID, dhcpOptionID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}
