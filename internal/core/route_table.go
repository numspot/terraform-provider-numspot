package core

import (
	"context"
	"strings"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

// ReadRouteTables reads route tables.
func ReadRouteTables(ctx context.Context, provider *client.NumSpotSDK, params *api.ReadRouteTablesParams) (*[]api.RouteTable, error) {
	res, err := provider.Client.ReadRouteTablesWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read route tables")
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read route tables")
	}

	return res.JSON200.Items, nil
}

// ReadRouteTable reads route table.
func ReadRouteTable(ctx context.Context, provider *client.NumSpotSDK, id string) (*api.RouteTable, error) {
	res, err := provider.Client.ReadRouteTablesByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read route tables by id")
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read route tables by id")
	}

	return res.JSON200, nil
}

// CreateRouteTable creates route table.
func CreateRouteTable(
	ctx context.Context,
	provider *client.NumSpotSDK,
	payload *api.CreateRouteTableJSONRequestBody,
	routes []api.Route,
	subnetIDs []string,
	tags []api.ResourceTag,
) (*api.RouteTable, error) {
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.SpaceID,
		*payload,
		provider.Client.CreateRouteTableWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to create resource")
	}

	createdID := *res.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to create resource")
		}
	}

	if len(routes) > 0 {
		if err = createRouteTableRoutes(ctx, provider, createdID, routes); err != nil {
			return nil, utils.WrapError(err, "failed to create resource")
		}
	}

	for _, subnetId := range subnetIDs {
		if err = linkRouteTable(ctx, provider, createdID, subnetId); err != nil {
			return nil, utils.WrapError(err, "failed to create resource")
		}
	}

	return ReadRouteTable(ctx, provider, createdID)
}

// DeleteRouteTable deletes route table.
func DeleteRouteTable(ctx context.Context, provider *client.NumSpotSDK, id string, links []string) error {
	for _, link := range links {
		if err := unlinkRouteTable(ctx, provider, id, link); err != nil {
			return utils.WrapError(err, "failed to create resource")
		}
	}
	return utils.WrapError(
		utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, provider.Client.DeleteRouteTableWithResponse),
		"failed to delete resource",
	)
}

// UpdateRouteTableRoutes updates route table routes.
func UpdateRouteTableRoutes(
	ctx context.Context,
	provider *client.NumSpotSDK,
	id string,
	stateRoutes []api.Route,
	planRoutes []api.Route,
) (*api.RouteTable, error) {
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

// UpdateRouteTableTags updates route table tags.
func UpdateRouteTableTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	id string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.RouteTable, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, id); err != nil {
		return nil, err
	}
	return ReadRouteTable(ctx, provider, id)
}

// UpdateRouteTableSubnets updates route table subnets.
func UpdateRouteTableSubnets(
	ctx context.Context,
	provider *client.NumSpotSDK,
	routeTableId string,
	desiredSubnetIDs []string,
) (*api.RouteTable, error) {
	current, err := ReadRouteTable(ctx, provider, routeTableId)
	if err != nil {
		return nil, err
	}

	currentSubnetToLink := buildSubnetToLinkMap(current)
	desiredSet := buildDesiredSubnetSet(desiredSubnetIDs)

	if err := unlinkRemovedRouteTableSubnets(ctx, provider, routeTableId, currentSubnetToLink, desiredSet); err != nil {
		return nil, err
	}

	if err := linkNewRouteTableSubnets(ctx, provider, routeTableId, desiredSubnetIDs, currentSubnetToLink); err != nil {
		return nil, err
	}

	return ReadRouteTable(ctx, provider, routeTableId)
}

func buildSubnetToLinkMap(current *api.RouteTable) map[string]string {
	currentSubnetToLink := make(map[string]string)
	if current.LinkRouteTables == nil {
		return currentSubnetToLink
	}

	for _, link := range *current.LinkRouteTables {
		if link.SubnetId != nil && link.Id != nil {
			currentSubnetToLink[*link.SubnetId] = *link.Id
		}
	}

	return currentSubnetToLink
}

func buildDesiredSubnetSet(desiredSubnetIDs []string) map[string]bool {
	desiredSet := make(map[string]bool, len(desiredSubnetIDs))
	for _, subnetId := range desiredSubnetIDs {
		desiredSet[subnetId] = true
	}

	return desiredSet
}

func unlinkRemovedRouteTableSubnets(
	ctx context.Context,
	provider *client.NumSpotSDK,
	routeTableId string,
	currentSubnetToLink map[string]string,
	desiredSet map[string]bool,
) error {
	for subnetId, linkId := range currentSubnetToLink {
		if desiredSet[subnetId] {
			continue
		}

		if err := unlinkRouteTable(ctx, provider, routeTableId, linkId); err != nil {
			return err
		}
	}

	return nil
}

func linkNewRouteTableSubnets(
	ctx context.Context,
	provider *client.NumSpotSDK,
	routeTableId string,
	desiredSubnetIDs []string,
	currentSubnetToLink map[string]string,
) error {
	for _, subnetId := range desiredSubnetIDs {
		if _, exists := currentSubnetToLink[subnetId]; exists {
			continue
		}

		if err := linkRouteTable(ctx, provider, routeTableId, subnetId); err != nil {
			return err
		}
	}

	return nil
}

func createRouteTableRoutes(ctx context.Context, provider *client.NumSpotSDK, routeTableId string, routes []api.Route) error {
	for _, r := range routes {
		payload := api.CreateRoute{
			DestinationIpRange: *r.DestinationIpRange,
			GatewayId:          r.GatewayId,
			NatGatewayId:       r.NatGatewayId,
			NicId:              r.NicId,
			VmId:               r.VmId,
			VpcPeeringId:       r.VpcPeeringId,
		}

		res, err := provider.Client.CreateRouteWithResponse(ctx, provider.SpaceID, routeTableId, payload)
		if err != nil {
			return utils.WrapError(err, "failed to create route")
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			return utils.WrapError(err, "failed to create route")
		}
	}
	return nil
}

func deleteRouteTableRoutes(ctx context.Context, provider *client.NumSpotSDK, routeTableId string, routes []api.Route) error {
	for _, r := range routes {
		payload := api.DeleteRoute{}
		if r.DestinationIpRange != nil {
			payload.DestinationIpRange = *r.DestinationIpRange
		}
		res, err := provider.Client.DeleteRouteWithResponse(ctx, provider.SpaceID, routeTableId, payload)
		if err != nil {
			return utils.WrapError(err, "failed to delete route")
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			return utils.WrapError(err, "failed to delete route")
		}
	}
	return nil
}

func removeLocalRouteFromRoutes(routes []api.Route) []api.Route {
	arr := make([]api.Route, 0)
	for _, route := range routes {
		if route.GatewayId != nil && !strings.EqualFold(*route.GatewayId, "local") {
			arr = append(arr, route)
		}
	}

	return arr
}

func linkRouteTable(ctx context.Context, provider *client.NumSpotSDK, routeTableId, subnetId string) error {
	res, err := provider.Client.LinkRouteTableWithResponse(ctx, provider.SpaceID, routeTableId, api.LinkRouteTableJSONRequestBody{SubnetId: subnetId})
	if err != nil {
		return utils.WrapError(err, "failed to link route table")
	}
	return utils.WrapError(utils.ParseHTTPError(res.Body, res.StatusCode()), "failed to parse HTTP error response")
}

func unlinkRouteTable(ctx context.Context, provider *client.NumSpotSDK, routeTableId, linkRouteTableId string) error {
	res, err := provider.Client.UnlinkRouteTableWithResponse(
		ctx,
		provider.SpaceID,
		routeTableId,
		api.UnlinkRouteTableJSONRequestBody{LinkRouteTableId: linkRouteTableId})
	if err != nil {
		return utils.WrapError(err, "failed to unlink route table")
	}
	return utils.WrapError(utils.ParseHTTPError(res.Body, res.StatusCode()), "failed to parse HTTP error response")
}
