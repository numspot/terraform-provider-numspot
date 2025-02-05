package core

import (
	"context"
	"strings"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func ReadRouteTables(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadRouteTablesParams) (*[]numspot.RouteTable, error) {
	res, err := provider.Client.ReadRouteTablesWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200.Items, nil
}

func ReadRouteTable(ctx context.Context, provider *client.NumSpotSDK, id string) (*numspot.RouteTable, error) {
	res, err := provider.Client.ReadRouteTablesByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func CreateRouteTable(
	ctx context.Context,
	provider *client.NumSpotSDK,
	payload numspot.CreateRouteTableJSONRequestBody,
	tags []numspot.ResourceTag,
	routes []numspot.Route,
	subnetID *string,
) (*numspot.RouteTable, error) {
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.SpaceID,
		payload,
		provider.Client.CreateRouteTableWithResponse)
	if err != nil {
		return nil, err
	}

	createdID := *res.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdID, tags); err != nil {
			return nil, err
		}
	}

	if len(routes) > 0 {
		if err = createRouteTableRoutes(ctx, provider, createdID, routes); err != nil {
			return nil, err
		}
	}

	if subnetID != nil {
		if err = linkRouteTable(ctx, provider, createdID, *subnetID); err != nil {
			return nil, err
		}
	}

	return ReadRouteTable(ctx, provider, createdID)
}

func DeleteRouteTable(ctx context.Context, provider *client.NumSpotSDK, id string, links []string) error {
	for _, link := range links {
		if err := unlinkRouteTable(ctx, provider, id, link); err != nil {
			return err
		}
	}
	return utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, provider.Client.DeleteRouteTableWithResponse)
}

func UpdateRouteTableRoutes(
	ctx context.Context,
	provider *client.NumSpotSDK,
	id string,
	stateRoutes []numspot.Route,
	planRoutes []numspot.Route,
) (*numspot.RouteTable, error) {
	stateRoutesWithoutLocal := removeLocalRouteFromRoutes(stateRoutes)
	toCreate, toDelete := utils.DiffComparable(stateRoutesWithoutLocal, planRoutes)
	if err := createRouteTableRoutes(ctx, provider, id, toCreate); err != nil {
		return nil, err
	}

	if err := deleteRouteTableRoutes(ctx, provider, id, toDelete); err != nil {
		return nil, err
	}

	return ReadRouteTable(ctx, provider, id)
}

func UpdateRouteTableTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	id string,
	stateTags []numspot.ResourceTag,
	planTags []numspot.ResourceTag,
) (*numspot.RouteTable, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, id); err != nil {
		return nil, err
	}
	return ReadRouteTable(ctx, provider, id)
}

func createRouteTableRoutes(ctx context.Context, provider *client.NumSpotSDK, routeTableId string, routes []numspot.Route) error {
	for _, r := range routes {
		payload := numspot.CreateRoute{
			GatewayId:    r.GatewayId,
			NatGatewayId: r.NatGatewayId,
			NicId:        r.NicId,
			VmId:         r.VmId,
			VpcPeeringId: r.VpcPeeringId,
		}
		if r.DestinationIpRange != nil {
			payload.DestinationIpRange = *r.DestinationIpRange
		}

		res, err := provider.Client.CreateRouteWithResponse(ctx, provider.SpaceID, routeTableId, payload)
		if err != nil {
			return err
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			return err
		}
	}
	return nil
}

func deleteRouteTableRoutes(ctx context.Context, provider *client.NumSpotSDK, routeTableId string, routes []numspot.Route) error {
	for _, r := range routes {
		payload := numspot.DeleteRoute{}
		if r.DestinationIpRange != nil {
			payload.DestinationIpRange = *r.DestinationIpRange
		}
		res, err := provider.Client.DeleteRouteWithResponse(ctx, provider.SpaceID, routeTableId, payload)
		if err != nil {
			return err
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			return err
		}
	}
	return nil
}

func removeLocalRouteFromRoutes(routes []numspot.Route) []numspot.Route {
	arr := make([]numspot.Route, 0)
	for _, route := range routes {
		if route.GatewayId != nil && !strings.EqualFold(*route.GatewayId, "local") {
			arr = append(arr, route)
		}
	}

	return arr
}

func linkRouteTable(ctx context.Context, provider *client.NumSpotSDK, routeTableId, subnetId string) error {
	res, err := provider.Client.LinkRouteTableWithResponse(ctx, provider.SpaceID, routeTableId, numspot.LinkRouteTableJSONRequestBody{SubnetId: subnetId})
	if err != nil {
		return err
	}
	return utils.ParseHTTPError(res.Body, res.StatusCode())
}

func unlinkRouteTable(ctx context.Context, provider *client.NumSpotSDK, routeTableId, linkRouteTableId string) error {
	res, err := provider.Client.UnlinkRouteTableWithResponse(
		ctx,
		provider.SpaceID,
		routeTableId,
		numspot.UnlinkRouteTableJSONRequestBody{LinkRouteTableId: linkRouteTableId})
	if err != nil {
		return err
	}
	return utils.ParseHTTPError(res.Body, res.StatusCode())
}
