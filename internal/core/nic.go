package core

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateNic(ctx context.Context, provider *client.NumSpotSDK, numSpotNicCreate numspot.CreateNicJSONRequestBody, tags []numspot.ResourceTag, linkNicBody *numspot.LinkNicJSONRequestBody) (*numspot.Nic, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreateResponse *numspot.CreateNicResponse
	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, numSpotNicCreate, numspotClient.CreateNicWithResponse); err != nil {
		return nil, err
	}

	nicID := *retryCreateResponse.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, nicID, tags); err != nil {
			return nil, err
		}
	}

	nic, err := ReadNicWithID(ctx, provider, nicID)
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

func UpdateNicTags(ctx context.Context, provider *client.NumSpotSDK, nicID string, stateTags, planTags []numspot.ResourceTag) (*numspot.Nic, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, nicID); err != nil {
		return nil, err
	}
	return ReadNicWithID(ctx, provider, nicID)
}

func linkNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, linkNicBody numspot.LinkNicJSONRequestBody) (*numspot.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.LinkNicWithResponse(ctx, provider.SpaceID, nicID, linkNicBody)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		return nil, apiError
	}

	nic, err := RetryReadLinkNic(ctx, provider, nicID, []string{attaching}, []string{attached})
	if err != nil {
		return nil, err
	}

	return nic, nil
}

func unlinkNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, linkNicBody numspot.UnlinkNicJSONRequestBody) (*numspot.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.UnlinkNicWithResponse(ctx, provider.SpaceID, nicID, linkNicBody)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusNoContent {
		apiError := utils.HandleError(res.Body)
		return nil, apiError
	}

	nic, err := RetryReadNic(ctx, provider, nicID, []string{inUse, detaching}, []string{available})
	if err != nil {
		return nil, err
	}

	return nic, nil
}

func UpdateNicLink(ctx context.Context, provider *client.NumSpotSDK, nicID string, stateUnlinkNic *numspot.UnlinkNicJSONRequestBody, planLinkNic *numspot.LinkNicJSONRequestBody) (*numspot.Nic, error) {
	var nic *numspot.Nic
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

func UpdateNicAttributes(ctx context.Context, provider *client.NumSpotSDK, numSpotNicUpdate numspot.UpdateNicJSONRequestBody, nicID string) (*numspot.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.UpdateNicWithResponse(ctx, provider.SpaceID, nicID, numSpotNicUpdate)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		return nil, apiError
	}

	return res.JSON200, nil
}

func DeleteNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, unlinkNicBody *numspot.UnlinkNicJSONRequestBody) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	if unlinkNicBody != nil {
		_, _ = unlinkNic(ctx, provider, nicID, *unlinkNicBody)
		// Error not handled, we try to delete internet gateway anyway
	}

	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, nicID, numspotClient.DeleteNicWithResponse)
}

func RetryReadLinkNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, startState, targetState []string) (*numspot.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	createStateConf := &retry.StateChangeConf{
		Pending: startState,
		Target:  targetState,
		Refresh: func() (interface{}, string, error) {
			resp, err := numspotClient.ReadNicsByIdWithResponse(ctx, provider.SpaceID, nicID)
			if err != nil {
				return nil, "", err
			}

			if resp.StatusCode() != http.StatusOK {
				apiError := utils.HandleError(resp.Body)
				return nil, "", apiError
			}

			if resp != nil && resp.JSON200 != nil && resp.JSON200.LinkNic != nil && resp.JSON200.State != nil {
				return resp.JSON200, *resp.JSON200.LinkNic.State, nil
			} else {
				return nil, "", fmt.Errorf("error while reading operation. No 'LinkNic.State' field found in response")
			}
		},
		Timeout: utils.TfRequestRetryTimeout,
		Delay:   utils.ParseRetryBackoff(),
	}

	read, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	numSpotNic, assert := read.(*numspot.Nic)
	if !assert {
		return nil, fmt.Errorf("invalid nic assertion %s", nicID)
	}
	return numSpotNic, err
}

func RetryReadNic(ctx context.Context, provider *client.NumSpotSDK, nicID string, startState, targetState []string) (*numspot.Nic, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, nicID, provider.SpaceID, startState, targetState, numspotClient.ReadNicsByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotNic, assert := read.(*numspot.Nic)
	if !assert {
		return nil, fmt.Errorf("invalid nic assertion %s", nicID)
	}
	return numSpotNic, err
}

func ReadNicWithID(ctx context.Context, provider *client.NumSpotSDK, nicID string) (numSpotNic *numspot.Nic, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadNic, err := numspotClient.ReadNicsByIdWithResponse(ctx, provider.SpaceID, nicID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadNic.Body, numSpotReadNic.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadNic.JSON200, err
}

func ReadNicsWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadNicsParams) (numSpotNic *[]numspot.Nic, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadNic, err := numspotClient.ReadNicsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadNic.Body, numSpotReadNic.StatusCode()); err != nil {
		return nil, err
	}
	if numSpotReadNic.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of nic but got nil")
	}

	return numSpotReadNic.JSON200.Items, err
}
