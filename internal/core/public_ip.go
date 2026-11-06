package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// errExpectedPublicIPList is returned when a public IP list response is
// unexpectedly nil.
var errExpectedPublicIPList = errors.New("HTTP call failed : expected a list of public Ips but got nil")

// CreatePublicIp creates public IP.
func CreatePublicIp(
	ctx context.Context,
	provider *client.NumspotSDK,
	vmId, nicId string,
	tags []api.ResourceTag,
) (numspotPublicIp *api.PublicIp, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreatePublicIpResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailable(ctx, spaceID, numspotClient.CreatePublicIpWithResponse); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	if retryCreate.JSON201 == nil || retryCreate.JSON201.Id == nil {
		return nil, fmt.Errorf("%w (HTTP %d)", errMissingCreateID, retryCreate.StatusCode())
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdId, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	// Attach the public IP to a VM or NIC if their IDs are provided:
	// Note: According to the resource schema, vmId and nicId cannot be set simultaneously.
	// This constraint is enforced by the stringvalidator.ConflictsWith function.
	if vmId != "" || nicId != "" {
		// Call Link publicIP
		if err = linkPublicIP(ctx, provider, createdId, vmId, nicId); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	return ReadPublicIp(ctx, provider, createdId)
}

// UpdatePublicIpTarget moves the public IP link to the target VM or NIC given
// in the plan. The current link is read from the API first, so a stale state
// (e.g. the linked VM was replaced during the same apply and its destruction
// already removed the link) does not fail the update. When the IP is already
// linked to the planned target, no API call mutating the IP is performed. The
// IP resource itself is never destroyed here: it keeps its address (SUPN2-643).
func UpdatePublicIpTarget(
	ctx context.Context,
	provider *client.NumspotSDK,
	publicIpID string,
	planVmId, planNicId string,
) (*api.PublicIp, error) {
	current, err := ReadPublicIp(ctx, provider, publicIpID)
	if err != nil {
		return nil, err
	}

	// Already linked to the planned target (covers both linked and unlinked).
	if actualVmId, actualNicId := publicIPTarget(current); actualVmId == planVmId && actualNicId == planNicId {
		return current, nil
	}

	if err := relinkPublicIP(ctx, provider, publicIpID, current.LinkPublicIpId, planVmId, planNicId); err != nil {
		return nil, err
	}

	return ReadPublicIp(ctx, provider, publicIpID)
}

// publicIPTarget returns the (vmId, nicId) pair the public IP is currently
// linked to; both are empty when the IP is unlinked.
func publicIPTarget(ip *api.PublicIp) (vmId, nicId string) {
	if ip.VmId != nil {
		vmId = *ip.VmId
	}
	if ip.NicId != nil {
		nicId = *ip.NicId
	}
	return vmId, nicId
}

// relinkPublicIP unlinks the current association (if any) then links the IP to
// the planned VM or NIC (if any).
func relinkPublicIP(
	ctx context.Context,
	provider *client.NumspotSDK,
	publicIpID string,
	currentLinkID *string,
	planVmId, planNicId string,
) error {
	if currentLinkID != nil && *currentLinkID != "" {
		if err := unlinkPublicIP(ctx, provider, publicIpID, *currentLinkID); err != nil {
			return err
		}
	}
	if planVmId == "" && planNicId == "" {
		return nil
	}
	return linkPublicIP(ctx, provider, publicIpID, planVmId, planNicId)
}

func linkPublicIP(ctx context.Context, provider *client.NumspotSDK, publicIpId, vmId, nicId string) error {
	var payload api.LinkPublicIpJSONRequestBody
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	if vmId != "" {
		payload = api.LinkPublicIpJSONRequestBody{VmId: &vmId}
	} else {
		payload = api.LinkPublicIpJSONRequestBody{NicId: &nicId}
	}
	linkPublicIPResponse, err := numspotClient.LinkPublicIpWithResponse(ctx, provider.SpaceID, publicIpId, payload)
	if err != nil {
		return utils.WrapError(err, "failed to link public ip")
	}
	if err = utils.ParseHTTPError(linkPublicIPResponse.Body, linkPublicIPResponse.StatusCode()); err != nil {
		return utils.WrapError(err, "failed to link public ip")
	}

	return nil
}

func unlinkPublicIP(ctx context.Context, provider *client.NumspotSDK, publicIpID, linkPublicIpID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	if _, err = utils.RetryDeleteUntilWithBody(
		ctx,
		provider.SpaceID,
		publicIpID,
		api.UnlinkPublicIpJSONRequestBody{LinkPublicIpId: &linkPublicIpID},
		numspotClient.UnlinkPublicIpWithResponse,
	); err != nil {
		return utils.WrapError(err, "failed to unlink public ip")
	}

	return nil
}

// UpdatePublicIpTags updates public IP tags.
func UpdatePublicIpTags(
	ctx context.Context,
	provider *client.NumspotSDK,
	publicIpID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.PublicIp, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, publicIpID); err != nil {
		return nil, utils.WrapError(err, "failed to link public ip")
	}
	return ReadPublicIp(ctx, provider, publicIpID)
}

// DeletePublicIp deletes public IP.
func DeletePublicIp(ctx context.Context, provider *client.NumspotSDK, publicIpID, linkPublicIpID string) error {
	spaceID := provider.SpaceID

	if linkPublicIpID != "" {
		if err := unlinkPublicIP(ctx, provider, publicIpID, linkPublicIpID); err != nil {
			return err
		}
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, spaceID, publicIpID, numspotClient.DeletePublicIpWithResponse),
		"failed to delete resource",
	)
}

// ReadPublicIp reads public IP.
func ReadPublicIp(ctx context.Context, provider *client.NumspotSDK, publicIpID string) (*api.PublicIp, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotPublicIp, err := numspotClient.ReadPublicIpsByIdWithResponse(ctx, provider.SpaceID, publicIpID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read public ips by id")
	}

	if err = utils.ParseHTTPError(numspotPublicIp.Body, numspotPublicIp.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read public ips by id")
	}

	return numspotPublicIp.JSON200, nil
}

// ReadPublicIpsWithParams reads public IPs with params.
func ReadPublicIpsWithParams(
	ctx context.Context,
	provider *client.NumspotSDK,
	params *api.ReadPublicIpsParams,
) (numspotPublicIp *[]api.PublicIp, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadPublicIp, err := numspotClient.ReadPublicIpsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read public ips")
	}
	if err = utils.ParseHTTPError(numspotReadPublicIp.Body, numspotReadPublicIp.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read public ips")
	}
	if numspotReadPublicIp.JSON200.Items == nil {
		return nil, errExpectedPublicIPList
	}

	return numspotReadPublicIp.JSON200.Items, utils.WrapError(err, "failed to read public ips")
}
