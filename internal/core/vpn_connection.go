package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	vpnConnectionPendingStates = []string{pending}
	vpnConnectionTargetStates  = []string{available}
)

// Adapted from utils.RetryCreateUntilResourceAvailableWithBody
func retryCreateVpnCondition(
	ctx context.Context,
	provider *client.NumSpotSDK,
	spaceID api.SpaceId,
	body api.CreateVPNConnectionJSONRequestBody,
) (*api.CreateVPNConnectionResponse, error) {
	var res *api.CreateVPNConnectionResponse

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	retryError := retry.RetryContext(ctx, utils.TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = numspotClient.CreateVPNConnectionWithResponse(ctx, spaceID, body)
		if err != nil {
			return retry.NonRetryableError(err)
		}
		if slices.Contains(utils.StatusCodeStopRetryOnCreate, res.StatusCode()) {
			// If a VPN Connection using the same client/virtual gateway got deleted recently, the create call will return the ID of the deleted VPN Connection
			// In this case, we retry to create the VPN connection
			time.Sleep(5 * time.Second)
			_, err := RetryReadVpnConnection(ctx, provider, (*res.JSON201).Id.String())
			if err != nil {
				return retry.RetryableError(fmt.Errorf("error : retry timeout reached (%v). Error message : %v", utils.TfRequestRetryTimeout, err))
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
	numSpotVpnConnectionCreate api.CreateVPNConnectionJSONRequestBody,
	routes []api.CreateVPNConnectionRoute,
) (numSpotVpnConnection *api.VPNConnection, err error) {
	spaceID := provider.SpaceID

	var retryCreate *api.CreateVPNConnectionResponse
	if retryCreate, err = retryCreateVpnCondition(ctx, provider, spaceID, numSpotVpnConnectionCreate); err != nil {
		return nil, err
	}

	createdId := (*retryCreate.JSON201).Id

	if len(routes) > 0 {
		err = addRoutes(ctx, provider, createdId, routes)
		if err != nil {
			return nil, err
		}
	}

	return RetryReadVpnConnection(ctx, provider, createdId.String())
}

func UpdateVpnConnectionRoutes(ctx context.Context, provider *client.NumSpotSDK, routesToDelete []api.DeleteVPNConnectionRoute, routesToCreate []api.CreateVPNConnectionRoute, vpnConnectionID api.ResourceIdentifier) (*api.VPNConnection, error) {
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

func DeleteVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID api.ResourceIdentifier) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, vpnConnectionID, numspotClient.DeleteVPNConnectionWithResponse)
}

func ReadVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID api.ResourceIdentifier) (*api.VPNConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotVpnConnection, err := numspotClient.ReadVPNConnectionWithResponse(ctx, provider.SpaceID, vpnConnectionID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotVpnConnection.Body, numSpotVpnConnection.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotVpnConnection.JSON200, nil
}

func RetryReadVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID string) (*api.VPNConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, uuid.MustParse(vpnConnectionID), provider.SpaceID, vpnConnectionPendingStates, vpnConnectionTargetStates, numspotClient.ReadVPNConnectionWithResponse)
	if err != nil {
		return nil, err
	}

	vpnConnection, assert := read.(*api.VPNConnection)
	if !assert {
		return nil, fmt.Errorf("invalid vpn connection assertion %s", vpnConnectionID)
	}
	return vpnConnection, err
}

func ReadVpnConnectionsWithParams(ctx context.Context, provider *client.NumSpotSDK) (numSpotVpnConnection []api.VPNConnection, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadVpnConnection, err := numspotClient.ListVPNConnectionsWithResponse(ctx, provider.SpaceID)
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

func addRoutes(ctx context.Context, provider *client.NumSpotSDK, vpnID api.ResourceIdentifier, routes []api.CreateVPNConnectionRoute) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	for _, route := range routes {
		res, err := numspotClient.CreateVPNConnectionRouteWithResponse(ctx, provider.SpaceID, vpnID, route)
		if err != nil {
			return err
		}

		apiError := utils.HandleError(res.Body)
		if apiError.Error() != "" {
			return fmt.Errorf("error while adding route %v : %v", route, apiError)
		}
	}
	return nil
}

func deleteRoutes(ctx context.Context, provider *client.NumSpotSDK, vpnID api.ResourceIdentifier, routes []api.DeleteVPNConnectionRoute) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	for _, route := range routes {
		res, err := numspotClient.DeleteVPNConnectionRouteWithResponse(ctx, provider.SpaceID, vpnID, route)
		if err != nil {
			return err
		}

		apiError := utils.HandleError(res.Body)
		if apiError.Error() != "" {
			return fmt.Errorf("error while deleting route %v : %v", route, apiError)
		}
	}
	return nil
}
