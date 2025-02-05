package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	vpnConnectionPendingStates = []string{pending}
	vpnConnectionTargetStates  = []string{available}
)

// Adapted from utils.RetryCreateUntilResourceAvailableWithBody
func retryCreateVpnCondition(
	ctx context.Context,
	provider *client.NumSpotSDK,
	spaceID numspot.SpaceId,
	body numspot.CreateVpnConnectionJSONRequestBody,
) (*numspot.CreateVpnConnectionResponse, error) {
	var res *numspot.CreateVpnConnectionResponse

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	retryError := retry.RetryContext(ctx, utils.TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = numspotClient.CreateVpnConnectionWithResponse(ctx, spaceID, body)
		if err != nil {
			return retry.NonRetryableError(err)
		}
		if slices.Contains(utils.StatusCodeStopRetryOnCreate, res.StatusCode()) {
			// If a VPN Connection using the same client/virtual gateway got deleted recently, the create call will return the ID of the deleted VPN Connection
			// In this case, we retry to create the VPN connection
			time.Sleep(5 * time.Second)
			_, err := RetryReadVpnConnection(ctx, provider, *res.JSON201.Id)
			if err != nil {
				return retry.RetryableError(fmt.Errorf("error : retry timeout reached (%v). Error message : %v.", utils.TfRequestRetryTimeout, err))
			} else {
				return nil
			}
		} else {
			errorMessage, err := utils.GetErrorMessage(res)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("error : got http status code %v but failed to parse error message. Reason : %v", res.StatusCode(), err))
			}

			if slices.Contains(utils.StatusCodeRetryOnCreate, res.StatusCode()) {
				time.Sleep(utils.ParseRetryBackoff())
				return retry.RetryableError(fmt.Errorf("error : retry timeout reached (%v). Error message : %v", utils.TfRequestRetryTimeout, errorMessage))
			} else {
				return retry.NonRetryableError(errors.New(errorMessage))
			}
		}
	})

	return res, retryError
}

func CreateVpnConnection(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotVpnConnectionCreate numspot.CreateVpnConnectionJSONRequestBody,
	vpnOptions numspot.VpnOptionsToUpdate,
	routes []numspot.CreateVpnConnectionRoute,
	tags []numspot.ResourceTag,
) (numSpotVpnConnection *numspot.VpnConnection, err error) {
	spaceID := provider.SpaceID

	var retryCreate *numspot.CreateVpnConnectionResponse
	if retryCreate, err = retryCreateVpnCondition(ctx, provider, spaceID, numSpotVpnConnectionCreate); err != nil {
		return nil, err
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdId, tags); err != nil {
			return nil, err
		}
	}

	if len(routes) > 0 {
		err = addRoutes(ctx, provider, createdId, routes)
		if err != nil {
			return nil, err
		}
	}

	_, err = updateVPNOptions(ctx, provider, createdId, vpnOptions)
	if err != nil {
		return nil, err
	}
	return RetryReadVpnConnection(ctx, provider, createdId)
}

func UpdateVpnConnectionTags(ctx context.Context, provider *client.NumSpotSDK, stateTags, planTags []numspot.ResourceTag, vpnConnectionID string) (*numspot.VpnConnection, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, vpnConnectionID); err != nil {
		return nil, err
	}
	return ReadVpnConnection(ctx, provider, vpnConnectionID)
}

func UpdateVpnConnectionAttributes(ctx context.Context, provider *client.NumSpotSDK, id string, vpnConnectionUpdate numspot.UpdateVpnConnection) (*numspot.VpnConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.UpdateVpnConnectionWithResponse(
		ctx,
		provider.SpaceID,
		id,
		vpnConnectionUpdate)
	if err != nil {
		return nil, err
	}

	err = utils.ParseHTTPError(res.Body, res.StatusCode())
	if err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func UpdateVpnConnectionRoutes(ctx context.Context, provider *client.NumSpotSDK, routesToDelete []numspot.DeleteVpnConnectionRoute, routesToCreate []numspot.CreateVpnConnectionRoute, vpnConnectionID string) (*numspot.VpnConnection, error) {
	err := deleteRoutes(ctx, provider, vpnConnectionID, routesToDelete)
	if err != nil {
		return nil, err
	}

	err = addRoutes(ctx, provider, vpnConnectionID, routesToCreate)
	if err != nil {
		return nil, err
	}

	return ReadVpnConnection(ctx, provider, vpnConnectionID)
}

func DeleteVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, vpnConnectionID, numspotClient.DeleteVpnConnectionWithResponse)
	if err != nil {
		return err
	}

	return nil
}

func ReadVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID string) (*numspot.VpnConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotVpnConnection, err := numspotClient.ReadVpnConnectionsByIdWithResponse(ctx, provider.SpaceID, vpnConnectionID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotVpnConnection.Body, numSpotVpnConnection.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotVpnConnection.JSON200, nil
}

func RetryReadVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID string) (*numspot.VpnConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, vpnConnectionID, provider.SpaceID, vpnConnectionPendingStates, vpnConnectionTargetStates, numspotClient.ReadVpnConnectionsByIdWithResponse)
	if err != nil {
		return nil, err
	}

	vpnConnection, assert := read.(*numspot.VpnConnection)
	if !assert {
		return nil, fmt.Errorf("invalid vpn connection assertion %s", vpnConnectionID)
	}
	return vpnConnection, err
}

func ReadVpnConnectionsWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadVpnConnectionsParams) (numSpotVpnConnection *[]numspot.VpnConnection, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadVpnConnection, err := numspotClient.ReadVpnConnectionsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadVpnConnection.Body, numSpotReadVpnConnection.StatusCode()); err != nil {
		return nil, err
	}
	if numSpotReadVpnConnection.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of public Ips but got nil")
	}

	return numSpotReadVpnConnection.JSON200.Items, err
}

func updateVPNOptions(ctx context.Context, provider *client.NumSpotSDK, id string, vpnOptions numspot.VpnOptionsToUpdate) (*numspot.VpnConnection, error) {
	return UpdateVpnConnectionAttributes(ctx, provider, id, numspot.UpdateVpnConnectionJSONRequestBody{
		VpnOptions: &vpnOptions,
	})
}

func addRoutes(ctx context.Context, provider *client.NumSpotSDK, vpnID string, routes []numspot.CreateVpnConnectionRoute) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	for _, route := range routes {
		res, err := numspotClient.CreateVpnConnectionRouteWithResponse(ctx, provider.SpaceID, vpnID, route)
		if err != nil {
			return err
		}

		apiError := utils.HandleError(res.Body)
		if apiError.Error() != "" {
			return fmt.Errorf("Error while adding route %v : %v", route, apiError)
		}
	}
	return nil
}

func deleteRoutes(ctx context.Context, provider *client.NumSpotSDK, vpnID string, routes []numspot.DeleteVpnConnectionRoute) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	for _, route := range routes {
		res, err := numspotClient.DeleteVpnConnectionRouteWithResponse(ctx, provider.SpaceID, vpnID, route)
		if err != nil {
			return err
		}

		apiError := utils.HandleError(res.Body)
		if apiError.Error() != "" {
			return fmt.Errorf("Error while deleting route %v : %v", route, apiError)
		}
	}
	return nil
}
