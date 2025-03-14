package core

import (
	"context"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

func CreateDHCPOptions(ctx context.Context, provider *client.NumSpotSDK, numSpotDHCPOptionsCreate numspot.CreateDhcpOptionsJSONRequestBody, tags []numspot.ResourceTag) (numSpotDHCPOptions *numspot.DhcpOptionsSet, err error) {
	spaceID := provider.SpaceID

	var retryCreate *numspot.CreateDhcpOptionsResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotDHCPOptionsCreate,
		numspotClient.CreateDhcpOptionsWithResponse); err != nil {
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

func UpdateDHCPOptionsTags(ctx context.Context, provider *client.NumSpotSDK, dhcpOptionsID string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.DhcpOptionsSet, error) {
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

func ReadDHCPOptions(ctx context.Context, provider *client.NumSpotSDK, dhcpOptions numspot.ReadDhcpOptionsParams) (*[]numspot.DhcpOptionsSet, error) {
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

func ReadDHCPOption(ctx context.Context, provider *client.NumSpotSDK, dhcpOptionID string) (*numspot.DhcpOptionsSet, error) {
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
