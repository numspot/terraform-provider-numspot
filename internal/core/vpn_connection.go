package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
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
	provider *client.NumspotSDK,
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
			return confirmCreatedVPNConnection(ctx, provider, res)
		}

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
		}

		return retry.NonRetryableError(fmt.Errorf("%w: %s", errUnexpectedVPNConnectionHTTPStatus, errorMessage))
	})

	return res, utils.WrapError(retryError, "failed to retry operation")
}

// confirmCreatedVPNConnection verifies that a create "stop-retry" response
// really produced a usable VPN connection. If a VPN Connection using the same
// client/virtual gateway got deleted recently, the create call will return the
// ID of the deleted VPN Connection: in that case the create is retried.
func confirmCreatedVPNConnection(
	ctx context.Context,
	provider *client.NumspotSDK,
	res *api.CreateVPNConnectionResponse,
) *retry.RetryError {
	if res.JSON201 == nil {
		return retry.NonRetryableError(
			fmt.Errorf("%w (HTTP %d)", errMissingCreateID, res.StatusCode()),
		)
	}

	time.Sleep(5 * time.Second)
	if _, err := RetryReadVpnConnection(ctx, provider, res.JSON201.Id.String()); err != nil {
		return retry.RetryableError(
			fmt.Errorf("%w (%v). Error message : %v", errVPNConnectionRetryTimeout, utils.TfRequestRetryTimeout, err),
		)
	}

	return nil
}

// CreateVpnConnection creates VPN connection.
func CreateVpnConnection(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotVpnConnectionCreate *api.CreateVPNConnectionJSONRequestBody,
	routes []api.CreateVPNConnectionRoute,
) (numspotVpnConnection *api.VPNConnection, err error) {
	spaceID := provider.SpaceID

	var retryCreate *api.CreateVPNConnectionResponse
	if retryCreate, err = retryCreateVpnCondition(ctx, provider, spaceID, numspotVpnConnectionCreate); err != nil {
		return nil, utils.WrapError(err, "failed to get error message")
	}

	if retryCreate.JSON201 == nil {
		return nil, fmt.Errorf("%w (HTTP %d)", errMissingCreateID, retryCreate.StatusCode())
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
	provider *client.NumspotSDK,
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
func DeleteVpnConnection(ctx context.Context, provider *client.NumspotSDK, vpnConnectionID string) error {
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
func ReadVpnConnection(ctx context.Context, provider *client.NumspotSDK, vpnConnectionID string) (*api.VPNConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotVpnConnection, err := numspotClient.ReadVPNConnectionWithResponse(ctx, provider.SpaceID, utils.StringToUuid(vpnConnectionID))
	if err != nil {
		return nil, utils.WrapError(err, "failed to read vpn connection")
	}

	if err = utils.ParseHTTPError(numspotVpnConnection.Body, numspotVpnConnection.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read vpn connection")
	}

	return numspotVpnConnection.JSON200, nil
}

// RetryReadVpnConnection retries read VPN connection.
func RetryReadVpnConnection(ctx context.Context, provider *client.NumspotSDK, vpnConnectionID string) (*api.VPNConnection, error) {
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
func ReadVpnConnectionsWithParams(ctx context.Context, provider *client.NumspotSDK) (*[]api.VPNConnection, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadVpnConnection, err := numspotClient.ListVPNConnectionsWithResponse(ctx, provider.SpaceID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list vpn connections")
	}
	if err = utils.ParseHTTPError(numspotReadVpnConnection.Body, numspotReadVpnConnection.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to list vpn connections")
	}
	if numspotReadVpnConnection.JSON200.Items == nil {
		return nil, errExpectedVPNConnectionList
	}

	return &numspotReadVpnConnection.JSON200.Items, nil
}

func addRoutes(ctx context.Context, provider *client.NumspotSDK, vpnID api.ResourceIdentifier, routes []api.CreateVPNConnectionRoute) error {
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

func deleteRoutes(ctx context.Context, provider *client.NumspotSDK, vpnID api.ResourceIdentifier, routes []api.DeleteVPNConnectionRoute) error {
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
