package core

import (
	"context"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateVPC(ctx context.Context,
	provider *client.NumSpotSDK,
	payload numspot.CreateVpcJSONRequestBody,
	tags []numspot.ResourceTag,
	updatePayload *numspot.UpdateVpcJSONRequestBody) (*numspot.Vpc, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.SpaceID,
		payload,
		numspotClient.CreateVpcWithResponse)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}
	createdId := *res.JSON201.Id

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, createdId, tags); err != nil {
			return nil, err
		}
	}

	if updatePayload != nil {
		_, err := numspotClient.UpdateVpcWithResponse(ctx, provider.SpaceID, createdId, *updatePayload)
		if err != nil {
			return nil, err
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			return nil, err
		}
	}
	return ReadVPC(ctx, provider, createdId)

}
func ReadVPC(ctx context.Context, provider *client.NumSpotSDK, id string) (*numspot.Vpc, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.ReadVpcsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func UpdateVPCDHCP(ctx context.Context, provider *client.NumSpotSDK, id string, payload numspot.UpdateVpcJSONRequestBody) (*numspot.Vpc, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
}

func DeleteVPC(ctx context.Context, provider *client.NumSpotSDK, id string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, numspotClient.DeleteVpcWithResponse)
}
