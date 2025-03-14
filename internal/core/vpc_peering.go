package core

import (
	"context"
	"net/http"
	"time"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

func CreateVPCPeering(
	ctx context.Context,
	provider *client.NumSpotSDK,
	payload numspot.CreateVpcPeeringJSONRequestBody,
	tags []numspot.ResourceTag,
) (*numspot.VpcPeering, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	// In case of recreate sometimes due to Outscale internal caching mechanism,
	// sometimes the returned resource is the old one still deleting.
	// In order to bypass this limitation we call sleep func
	time.Sleep(2 * time.Second)

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.SpaceID,
		payload,
		numspotClient.CreateVpcPeeringWithResponse)
	if err != nil {
		return nil, err
	}

	vpcPeeringID := *res.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, vpcPeeringID, tags); err != nil {
			return nil, err
		}
	}

	// Retries can't success with retry_utils.RetryReadUntilStateValid because vpc_peering state is an object but a string.
	// Also, VPC Peering resource State is used to provided status of the peering process, not the creation process
	// So we do not need to implement specific retry process here.
	vpcPeering, err := ReadVPCPeering(ctx, provider, vpcPeeringID)
	if err != nil {
		return nil, err
	}

	return vpcPeering, nil
}

func ReadVPCPeering(ctx context.Context, provider *client.NumSpotSDK, vpcPeeringID string) (*numspot.VpcPeering, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.ReadVpcPeeringsByIdWithResponse(ctx, provider.SpaceID, vpcPeeringID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(res.Body, http.StatusOK); err != nil {
		return nil, err
	}
	return res.JSON200, nil
}

func ReadVPCPeerings(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadVpcPeeringsParams) (*[]numspot.VpcPeering, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.ReadVpcPeeringsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(res.Body, http.StatusOK); err != nil {
		return nil, err
	}

	return res.JSON200.Items, nil
}

func UpdateVPCPeeringTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	vpcPeeringID string,
	stateTags, planTags []numspot.ResourceTag,
) (*numspot.VpcPeering, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, vpcPeeringID); err != nil {
		return nil, err
	}

	vpcPeering, err := ReadVPCPeering(ctx, provider, vpcPeeringID)
	if err != nil {
		return nil, err
	}

	return vpcPeering, nil
}

func DeleteVPCPeering(ctx context.Context, provider *client.NumSpotSDK, vpcPeeringID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, vpcPeeringID, numspotClient.DeleteVpcPeeringWithResponse)
	if err != nil {
		return err
	}

	return nil
}
