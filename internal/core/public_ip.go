package core

import (
	"context"
	"fmt"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

func CreatePublicIp(ctx context.Context, provider *client.NumSpotSDK, tags []numspot.ResourceTag, vmId, nicId string) (numSpotPublicIp *numspot.PublicIp, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreatePublicIpResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailable(ctx, spaceID, numspotClient.CreatePublicIpWithResponse); err != nil {
		return nil, err
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdId, tags); err != nil {
			return nil, err
		}
	}

	// Attach the public IP to a VM or NIC if their IDs are provided:
	// Note: According to the resource schema, vmId and nicId cannot be set simultaneously.
	// This constraint is enforced by the stringvalidator.ConflictsWith function.
	if vmId != "" || nicId != "" {
		// Call Link publicIP
		if _, err = linkPublicIP(ctx, provider, createdId, vmId, nicId); err != nil {
			return nil, err
		}
	}

	return ReadPublicIp(ctx, provider, createdId)
}

func linkPublicIP(ctx context.Context, provider *client.NumSpotSDK, publicIpId, vmId, nicId string) (*string, error) {
	var payload numspot.LinkPublicIpJSONRequestBody
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	if vmId != "" {
		payload = numspot.LinkPublicIpJSONRequestBody{VmId: &vmId}
	} else {
		payload = numspot.LinkPublicIpJSONRequestBody{NicId: &nicId}
	}
	linkPublicIPResponse, err := numspotClient.LinkPublicIpWithResponse(ctx, provider.SpaceID, publicIpId, payload)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(linkPublicIPResponse.Body, linkPublicIPResponse.StatusCode()); err != nil {
		return nil, err
	}

	return linkPublicIPResponse.JSON200.LinkPublicIpId, nil
}

func UpdatePublicIpTags(ctx context.Context, provider *client.NumSpotSDK, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, publicIpID string) (*numspot.PublicIp, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, publicIpID); err != nil {
		return nil, err
	}
	return ReadPublicIp(ctx, provider, publicIpID)
}

func DeletePublicIp(ctx context.Context, provider *client.NumSpotSDK, publicIpID, linkPublicIpID string) error {
	spaceID := provider.SpaceID
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	if linkPublicIpID != "" {
		if _, err = utils.RetryDeleteUntilWithBody(ctx, spaceID, publicIpID, numspot.UnlinkPublicIpJSONRequestBody{LinkPublicIpId: &linkPublicIpID}, numspotClient.UnlinkPublicIpWithResponse); err != nil {
			return err
		}
	}

	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, publicIpID, numspotClient.DeletePublicIpWithResponse)
}

func ReadPublicIp(ctx context.Context, provider *client.NumSpotSDK, publicIpID string) (*numspot.PublicIp, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotPublicIp, err := numspotClient.ReadPublicIpsByIdWithResponse(ctx, provider.SpaceID, publicIpID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotPublicIp.Body, numSpotPublicIp.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotPublicIp.JSON200, nil
}

func ReadPublicIpsWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadPublicIpsParams) (numSpotPublicIp *[]numspot.PublicIp, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadPublicIp, err := numspotClient.ReadPublicIpsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadPublicIp.Body, numSpotReadPublicIp.StatusCode()); err != nil {
		return nil, err
	}
	if numSpotReadPublicIp.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of public Ips but got nil")
	}

	return numSpotReadPublicIp.JSON200.Items, err
}
