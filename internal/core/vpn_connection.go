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

	// errVPNConnectionRetryTimeout is returned when the VPN connection retry
	// loop exhausts its timeout.
	errVPNConnectionRetryTimeout = errors.New("error : retry timeout reached")
	// errVPNConnectionHTTPErrorParseFailed is returned when the error body of a
	// failed HTTP response cannot be parsed.
	errVPNConnectionHTTPErrorParseFailed = errors.New("error : got http status code")
	// errUnexpectedVPNConnectionHTTPStatus is the static sentinel for unexpected
	// HTTP status codes.
	errUnexpectedVPNConnectionHTTPStatus = errors.New("unexpected HTTP status")
	// errInvalidVPNConnectionAssertion is returned when a VPN connection read
	// result cannot be asserted to the expected type.
	errInvalidVPNConnectionAssertion = errors.New("invalid vpn connection assertion")
	// errExpectedVPNConnectionList is returned when a VPN connection list
	// response is unexpectedly nil.
	errExpectedVPNConnectionList = errors.New("HTTP call failed : expected a list of vpn connections but got nil")
	// errAddVPNConnectionRoute is returned when adding a VPN connection route
	// fails.
	errAddVPNConnectionRoute = errors.New("error while adding route")
	// errDeleteVPNConnectionRoute is returned when deleting a VPN connection
	// route fails.
	errDeleteVPNConnectionRoute = errors.New("error while deleting route")
)

// Adapted from utils.RetryCreateUntilResourceAvailableWithBody
func retryCreateVpnCondition(
	ctx context.Context,
	provider *client.NumSpotSDK,
	spaceID api.SpaceId,
	body *api.CreateVPNConnectionJSONRequestBody,
) (*api.CreateVPNConnectionResponse, error) {
	var res *api.CreateVPNConnectionResponse

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	retryError := retry.RetryContext(ctx, utils.TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = numspotClient.CreateVPNConnectionWithResponse(ctx, spaceID, *body)
		if err != nil {
			// Network-level failures (EOF, connection reset, timeout) are
			// transient and must be retried.
			return retry.RetryableError(utils.WrapError(err, "failed to create vpn connection"))
		}
		if slices.Contains(utils.StatusCodeStopRetryOnCreate, res.StatusCode()) {
			// If a VPN Connection using the same client/virtual gateway got deleted recently, the create call will return the ID of the deleted VPN Connection
			// In this case, we retry to create the VPN connection
			time.Sleep(5 * time.Second)
			_, err := RetryReadVpnConnection(ctx, provider, res.JSON201.Id.String())
			if err != nil {
				return retry.RetryableError(fmt.Errorf("%w (%v). Error message : %v", errVPNConnectionRetryTimeout, utils.TfRequestRetryTimeout, err))
			} else {
				return nil
			}
		} else {
			errorMessage, err := utils.GetErrorMessage(res)
			if err != nil {
				return retry.NonRetryableError(
					fmt.Errorf("%w %v but failed to parse error message. Reason : %v", errVPNConnectionHTTPErrorParseFailed, res.StatusCode(), err),
				)
			}

			if slices.Contains(utils.StatusCodeRetryOnCreate, res.StatusCode()) ||
				res.StatusCode() == 502 || res.StatusCode() == 503 || res.StatusCode() == 504 {
				time.Sleep(utils.ParseRetryBackoff())
				return retry.RetryableError(
					fmt.Errorf("%w (%v). Error message : %v", errVPNConnectionRetryTimeout, utils.TfRequestRetryTimeout, errorMessage),
				)
			} else {
				return retry.NonRetryableError(fmt.Errorf("%w: %s", errUnexpectedVPNConnectionHTTPStatus, errorMessage))
			}
		}
	})

	return res, utils.WrapError(retryError, "failed to retry operation")
}

// CreateVpnConnection creates VPN connection.
func CreateVpnConnection(
	ctx context.Context,
	provider *client.NumSpotSDK,
	numSpotVpnConnectionCreate *api.CreateVPNConnectionJSONRequestBody,
	routes []api.CreateVPNConnectionRoute,
) (numSpotVpnConnection *api.VPNConnection, err error) {
	spaceID := provider.SpaceID

	var retryCreate *api.CreateVPNConnectionResponse
	if retryCreate, err = retryCreateVpnCondition(ctx, provider, spaceID, numSpotVpnConnectionCreate); err != nil {
		return nil, utils.WrapError(err, "failed to get error message")
	}

	createdId := retryCreate.JSON201.Id

	if len(routes) > 0 {
		err = addRoutes(ctx, provider, createdId, routes)
		if err != nil {
			return nil, err
		}
	}

	return RetryReadVpnConnection(ctx, provider, createdId.String())
}

// UpdateVpnConnectionRoutes updates VPN connection routes.
func UpdateVpnConnectionRoutes(
	ctx context.Context,
	provider *client.NumSpotSDK,
	routesToDelete []api.DeleteVPNConnectionRoute,
	routesToCreate []api.CreateVPNConnectionRoute,
	vpnConnectionID string,
) (*api.VPNConnection, error) {
	id := utils.StringToUuid(vpnConnectionID)
	err := deleteRoutes(ctx, provider, id, routesToDelete)
	if err != nil {
		return nil, err
	}

	err = addRoutes(ctx, provider, id, routesToCreate)
	if err != nil {
		return nil, err
	}

	return ReadVpnConnection(ctx, provider, vpnConnectionID)
}

// DeleteVpnConnection deletes VPN connection.
func DeleteVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	err = utils.RetryDeleteUntilResourceAvailable(
		ctx,
		provider.SpaceID,
		utils.StringToUuid(vpnConnectionID),
		numspotClient.DeleteVPNConnectionWithResponse,
	)
	return utils.WrapError(err, "failed to delete resource")
}

// ReadVpnConnection reads VPN connection.
func ReadVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID string) (*api.VPNConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotVpnConnection, err := numspotClient.ReadVPNConnectionWithResponse(ctx, provider.SpaceID, utils.StringToUuid(vpnConnectionID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vpn connection")
	}

	if err = utils.ParseHTTPError(numSpotVpnConnection.Body, numSpotVpnConnection.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vpn connection")
	}

	return numSpotVpnConnection.JSON200, nil
}

// RetryReadVpnConnection retries read VPN connection.
func RetryReadVpnConnection(ctx context.Context, provider *client.NumSpotSDK, vpnConnectionID string) (*api.VPNConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		uuid.MustParse(vpnConnectionID),
		provider.SpaceID,
		vpnConnectionPendingStates,
		vpnConnectionTargetStates,
		numspotClient.ReadVPNConnectionWithResponse,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	vpnConnection, assert := read.(*api.VPNConnection)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidVPNConnectionAssertion, vpnConnectionID)
	}
	return vpnConnection, utils.WrapError(err, "failed to wait for resource state")
}

// ReadVpnConnectionsWithParams reads VPN connections with params.
func ReadVpnConnectionsWithParams(ctx context.Context, provider *client.NumSpotSDK) (*[]api.VPNConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadVpnConnection, err := numspotClient.ListVPNConnectionsWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list vpn connections")
	}
	if err = utils.ParseHTTPError(numSpotReadVpnConnection.Body, numSpotReadVpnConnection.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list vpn connections")
	}
	if numSpotReadVpnConnection.JSON200.Items == nil {
		return nil, errExpectedVPNConnectionList
	}

	return &numSpotReadVpnConnection.JSON200.Items, nil
}

func addRoutes(ctx context.Context, provider *client.NumSpotSDK, vpnID api.ResourceIdentifier, routes []api.CreateVPNConnectionRoute) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	for _, route := range routes {
		res, err := numspotClient.CreateVPNConnectionRouteWithResponse(ctx, provider.SpaceID, vpnID, route)
		if err != nil {
			return utils.WrapError(err, "failed to create vpn connection route")
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			return fmt.Errorf("%w %v : %v", errAddVPNConnectionRoute, route, err)
		}
	}
	return nil
}

func deleteRoutes(ctx context.Context, provider *client.NumSpotSDK, vpnID api.ResourceIdentifier, routes []api.DeleteVPNConnectionRoute) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	for _, route := range routes {
		res, err := numspotClient.DeleteVPNConnectionRouteWithResponse(ctx, provider.SpaceID, vpnID, route)
		if err != nil {
			return utils.WrapError(err, "failed to delete vpn connection route")
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			return fmt.Errorf("%w %v : %v", errDeleteVPNConnectionRoute, route, err)
		}
	}
	return nil
}
