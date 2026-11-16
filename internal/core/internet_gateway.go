package core

import (
	"context"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// CreateInternetGateway creates internet gateway.
func CreateInternetGateway(ctx context.Context, provider *client.NumspotSDK, vpcID string, tags []api.ResourceTag) (*api.InternetGateway, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreateResponse *api.CreateInternetGatewayResponse
	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailable(ctx, spaceID, numspotClient.CreateInternetGatewayWithResponse); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if retryCreateResponse.JSON201 == nil || retryCreateResponse.JSON201.Id == nil {
		return nil, fmt.Errorf("%w (HTTP %d)", errMissingCreateID, retryCreateResponse.StatusCode())
	}

	internetGatewayID := *retryCreateResponse.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, internetGatewayID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	if vpcID != "" {
		if err := linkInternetGatewayToVPC(ctx, provider, internetGatewayID, vpcID); err != nil {
			return nil, err
		}
	}

	internetGateway, err := ReadInternetGateway(ctx, provider, internetGatewayID)
	if err != nil {
		return nil, err
	}

	return internetGateway, nil
}

// linkInternetGatewayToVPC links an internet gateway to a VPC.
func linkInternetGatewayToVPC(ctx context.Context, provider *client.NumspotSDK, internetGatewayID, vpcID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	linkVPCResponse, err := numspotClient.LinkInternetGatewayWithResponse(ctx, provider.SpaceID, internetGatewayID,
		api.LinkInternetGatewayJSONRequestBody{
			VpcId: vpcID,
		},
	)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	return utils.WrapError(
		utils.ParseHTTPError(linkVPCResponse.Body, linkVPCResponse.StatusCode()),
		"failed to get API client",
	)
}

// UpdateInternetGatewayTags updates internet gateway tags.
func UpdateInternetGatewayTags(
	ctx context.Context,
	provider *client.NumspotSDK,
	internetGatewayID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.InternetGateway, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, internetGatewayID); err != nil {
		return nil, err
	}
	return ReadInternetGateway(ctx, provider, internetGatewayID)
}

// DeleteInternetGateway deletes internet gateway.
func DeleteInternetGateway(ctx context.Context, provider *client.NumspotSDK, internetGatewayID, vpcID string) (err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	if vpcID != "" {
		if _, err = utils.RetryUntilResourceAvailableWithBody(ctx, spaceID, internetGatewayID,
			api.UnlinkInternetGatewayJSONRequestBody{
				VpcId: vpcID,
			}, numspotClient.UnlinkInternetGatewayWithResponse); err != nil {
			return utils.WrapError(err, "failed to get API client")
		}
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, internetGatewayID, numspotClient.DeleteInternetGatewayWithResponse)
	if err != nil {
		return utils.WrapError(err, "failed to delete resource")
	}

	return nil
}

// ReadInternetGateway reads internet gateway.
func ReadInternetGateway(
	ctx context.Context,
	provider *client.NumspotSDK,
	internetGatewayID string,
) (numspotInternetGateway *api.InternetGateway, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadInternetGateway, err := numspotClient.ReadInternetGatewaysByIdWithResponse(ctx, provider.SpaceID, internetGatewayID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read internet gateways by id")
	}
	if err = utils.ParseHTTPError(numspotReadInternetGateway.Body, numspotReadInternetGateway.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read internet gateways by id")
	}

	return numspotReadInternetGateway.JSON200, utils.WrapError(err, "failed to read internet gateways by id")
}

// ReadInternetGatewaysWithParams reads internet gateways with params.
func ReadInternetGatewaysWithParams(
	ctx context.Context,
	provider *client.NumspotSDK,
	params *api.ReadInternetGatewaysParams,
) (numspotInternetGateway *[]api.InternetGateway, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadInternetGateway, err := numspotClient.ReadInternetGatewaysWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read internet gateways")
	}
	if err = utils.ParseHTTPError(numspotReadInternetGateway.Body, numspotReadInternetGateway.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read internet gateways")
	}

	return numspotReadInternetGateway.JSON200.Items, utils.WrapError(err, "failed to read internet gateways")
}
