package core

import (
	"context"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// CreateDHCPOptions creates DHCP options.
func CreateDHCPOptions(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotDHCPOptionsCreate *api.CreateDhcpOptionsJSONRequestBody,
	tags []api.ResourceTag,
) (numspotDHCPOptions *api.DhcpOptionsSet, err error) {
	spaceID := provider.SpaceID

	var retryCreate *api.CreateDhcpOptionsResponse
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*numspotDHCPOptionsCreate,
		numspotClient.CreateDhcpOptionsWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if retryCreate.JSON201 == nil || retryCreate.JSON201.Id == nil {
		return nil, fmt.Errorf("%w (HTTP %d)", errMissingCreateID, retryCreate.StatusCode())
	}

	dhcpOptionsID := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, dhcpOptionsID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	return ReadDHCPOption(ctx, provider, dhcpOptionsID)
}

// UpdateDHCPOptionsTags updates DHCP options tags.
func UpdateDHCPOptionsTags(
	ctx context.Context,
	provider *client.NumspotSDK,
	dhcpOptionsID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.DhcpOptionsSet, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, dhcpOptionsID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	return ReadDHCPOption(ctx, provider, dhcpOptionsID)
}

// DeleteDHCPOptions deletes DHCP options.
func DeleteDHCPOptions(ctx context.Context, provider *client.NumspotSDK, dhcpOptionsID string) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}
	if err := utils.RetryDeleteUntilResourceAvailable(ctx, spaceID, dhcpOptionsID,
		numspotClient.DeleteDhcpOptionsWithResponse); err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	return nil
}

// ReadDHCPOptions reads DHCP options.
func ReadDHCPOptions(ctx context.Context, provider *client.NumspotSDK, dhcpOptions *api.ReadDhcpOptionsParams) (*[]api.DhcpOptionsSet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := numspotClient.ReadDhcpOptionsWithResponse(ctx, provider.SpaceID, dhcpOptions)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read dhcp options")
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read dhcp options")
	}

	return read.JSON200.Items, nil
}

// ReadDHCPOption reads DHCP option.
func ReadDHCPOption(ctx context.Context, provider *client.NumspotSDK, dhcpOptionID string) (*api.DhcpOptionsSet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := numspotClient.ReadDhcpOptionsByIdWithResponse(ctx, provider.SpaceID, dhcpOptionID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read dhcp options by id")
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read dhcp options by id")
	}

	return read.JSON200, nil
}
