package core

import (
	"context"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateDHCPOptions(ctx context.Context, provider services.IProvider, numSpotDHCPOptionsCreate numspot.CreateDhcpOptionsJSONRequestBody, tags []numspot.ResourceTag) (numSpotDHCPOptions *numspot.DhcpOptionsSet, err error) {
	spaceID := provider.GetSpaceID()

	var retryCreate *numspot.CreateDhcpOptionsResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotDHCPOptionsCreate,
		provider.GetNumspotClient().CreateDhcpOptionsWithResponse); err != nil {
		return nil, err
	}

	dhcpOptionsID := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider.GetNumspotClient(), spaceID, dhcpOptionsID, tags); err != nil {
			return nil, err
		}
	}

	return ReadDHCPOption(ctx, provider, dhcpOptionsID)
}

func UpdateDHCPOptionsTags(ctx context.Context, provider services.IProvider, dhcpOptionsID string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.DhcpOptionsSet, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, dhcpOptionsID); err != nil {
		return nil, err
	}
	return ReadDHCPOption(ctx, provider, dhcpOptionsID)
}

func DeleteDHCPOptions(ctx context.Context, provider services.IProvider, dhcpOptionsID string) error {
	spaceID := provider.GetSpaceID()

	if err := utils.RetryDeleteUntilResourceAvailable(ctx, spaceID, dhcpOptionsID,
		provider.GetNumspotClient().DeleteDhcpOptionsWithResponse); err != nil {
		return err
	}

	return nil
}

func ReadDHCPOptions(ctx context.Context, provider services.IProvider, dhcpOptions numspot.ReadDhcpOptionsParams) (*numspot.ReadDhcpOptionsResponseSchema, error) {
	read, err := provider.GetNumspotClient().ReadDhcpOptionsWithResponse(ctx, provider.GetSpaceID(), &dhcpOptions)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}

func ReadDHCPOption(ctx context.Context, provider services.IProvider, dhcpOptionID string) (*numspot.DhcpOptionsSet, error) {
	read, err := provider.GetNumspotClient().ReadDhcpOptionsByIdWithResponse(ctx, provider.GetSpaceID(), dhcpOptionID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200, nil
}
