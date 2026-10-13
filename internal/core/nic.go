package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	// errNoLinkNicStateField is returned when a NIC response is missing the
	// LinkNic.State field.
	errNoLinkNicStateField = errors.New("error while reading operation. No 'LinkNic.State' field found in response")
	// errInvalidNicAssertion is returned when a NIC read result cannot be
	// asserted to the expected type.
	errInvalidNicAssertion = errors.New("invalid nic assertion")
	// errExpectedNicList is returned when a NIC list response is unexpectedly nil.
	errExpectedNicList = errors.New("HTTP call failed : expected a list of nic but got nil")
)

// CreateNic creates NIC.
func CreateNic(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotNicCreate *api.CreateNicJSONRequestBody,
	linkNicBody *api.LinkNicJSONRequestBody,
	tags []api.ResourceTag,
) (*api.Nic, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreateResponse *api.CreateNicResponse
	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*numSpotNicCreate,
		numspotClient.CreateNicWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	nicID := *retryCreateResponse.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, nicID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	nic, err := ReadNic(ctx, provider, nicID)
	if err != nil {
		return nil, err
	}

	if linkNicBody != nil {
		nic, err = linkNic(ctx, provider, nicID, *linkNicBody)
		if err != nil {
			return nil, err
		}
	}

	return nic, nil
}

// UpdateNicTags updates NIC tags.
func UpdateNicTags(ctx context.Context, provider *client.NumSpotSDK, nicID string, stateTags, planTags []api.ResourceTag) (*api.Nic, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, nicID); err != nil {
		return nil, err
	}
	return ReadNic(ctx, provider, nicID)
}

func linkNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, linkNicBody api.LinkNicJSONRequestBody) (*api.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.LinkNicWithResponse(ctx, provider.SpaceID, nicID, linkNicBody)
	if err != nil {
		return nil, utils.WrapError(err, "failed to link nic")
	}
	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		return nil, utils.WrapError(apiError, "failed to parse API error response")
	}

	nic, err := RetryReadLinkNic(ctx, provider, nicID, []string{attaching}, []string{attached})
	if err != nil {
		return nil, err
	}

	return nic, nil
}

func unlinkNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, linkNicBody api.UnlinkNicJSONRequestBody) (*api.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.UnlinkNicWithResponse(ctx, provider.SpaceID, nicID, linkNicBody)
	if err != nil {
		return nil, utils.WrapError(err, "failed to unlink nic")
	}

	if res.StatusCode() != http.StatusNoContent {
		apiError := utils.HandleError(res.Body)
		return nil, utils.WrapError(apiError, "failed to parse API error response")
	}

	nic, err := RetryReadNic(ctx, provider, nicID, []string{inUse, detaching}, []string{available})
	if err != nil {
		return nil, err
	}

	return nic, nil
}

// UpdateNicLink updates NIC link.
func UpdateNicLink(
	ctx context.Context,
	provider *client.NumSpotSDK,
	nicID string,
	stateUnlinkNic *api.UnlinkNicJSONRequestBody,
	planLinkNic *api.LinkNicJSONRequestBody,
) (*api.Nic, error) {
	var nic *api.Nic
	var err error
	if stateUnlinkNic != nil {
		nic, err = unlinkNic(ctx, provider, nicID, *stateUnlinkNic)
		if err != nil {
			return nil, err
		}
	}

	if planLinkNic != nil {
		nic, err = linkNic(ctx, provider, nicID, *planLinkNic)
		if err != nil {
			return nil, err
		}
	}

	return nic, nil
}

// UpdateNicAttributes updates NIC attributes.
func UpdateNicAttributes(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotNicUpdate *api.UpdateNicJSONRequestBody,
	nicID string,
) (*api.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.UpdateNicWithResponse(ctx, provider.SpaceID, nicID, *numSpotNicUpdate)
	if err != nil {
		return nil, utils.WrapError(err, "failed to update nic")
	}
	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		return nil, utils.WrapError(apiError, "failed to parse API error response")
	}

	return res.JSON200, nil
}

// DeleteNic deletes NIC.
func DeleteNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, unlinkNicBody *api.UnlinkNicJSONRequestBody) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	if unlinkNicBody != nil {
		_, _ = unlinkNic(ctx, provider, nicID, *unlinkNicBody)
		// Error not handled, we try to delete internet gateway anyway
	}

	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, nicID, numspotClient.DeleteNicWithResponse),
		"failed to delete resource",
	)
}

// RetryReadLinkNic retries read link NIC.
func RetryReadLinkNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, startState, targetState []string) (*api.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	createStateConf := &retry.StateChangeConf{
		Pending: startState,
		Target:  targetState,
		Refresh: func() (interface{}, string, error) {
			resp, err := numspotClient.ReadNicsByIdWithResponse(ctx, provider.SpaceID, nicID)
			if err != nil {
				return nil, "", utils.WrapError(err, "failed to read nics by id")
			}

			if resp.StatusCode() != http.StatusOK {
				apiError := utils.HandleError(resp.Body)
				return nil, "", utils.WrapError(apiError, "failed to parse API error response")
			}

			if resp != nil && resp.JSON200 != nil && resp.JSON200.LinkNic != nil && resp.JSON200.State != nil {
				return resp.JSON200, *resp.JSON200.LinkNic.State, nil
			} else {
				return nil, "", errNoLinkNicStateField
			}
		},
		Timeout: utils.TfRequestRetryTimeout,
		Delay:   utils.ParseRetryBackoff(),
	}

	read, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for state")
	}

	numSpotNic, assert := read.(*api.Nic)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidNicAssertion, nicID)
	}
	return numSpotNic, utils.WrapError(err, "failed to wait for state")
}

// RetryReadNic retries read NIC.
func RetryReadNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, startState, targetState []string) (*api.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(ctx, nicID, provider.SpaceID, startState, targetState, numspotClient.ReadNicsByIdWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numSpotNic, assert := read.(*api.Nic)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidNicAssertion, nicID)
	}
	return numSpotNic, utils.WrapError(err, "failed to wait for resource state")
}

// ReadNic reads NIC.
func ReadNic(ctx context.Context, provider *client.NumSpotSDK, nicID string) (numSpotNic *api.Nic, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadNic, err := numspotClient.ReadNicsByIdWithResponse(ctx, provider.SpaceID, nicID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read nics by id")
	}
	if err = utils.ParseHTTPError(numSpotReadNic.Body, numSpotReadNic.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read nics by id")
	}

	return numSpotReadNic.JSON200, utils.WrapError(err, "failed to read nics by id")
}

// ReadNicsWithParams reads NICs with params.
func ReadNicsWithParams(ctx context.Context, provider *client.NumSpotSDK, params *api.ReadNicsParams) (numSpotNic *[]api.Nic, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadNic, err := numspotClient.ReadNicsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read nics")
	}
	if err = utils.ParseHTTPError(numSpotReadNic.Body, numSpotReadNic.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read nics")
	}
	if numSpotReadNic.JSON200.Items == nil {
		return nil, errExpectedNicList
	}

	return numSpotReadNic.JSON200.Items, utils.WrapError(err, "failed to read nics")
}
