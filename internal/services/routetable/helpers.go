package routetable

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/routetable/resource_route_table"
	"terraform-provider-numspot/internal/utils"
)

// routesFromModel deserializes the routes set into the []api.Route the core
// create/update calls expect. The local route is never part of the user-managed
// routes, so it is not produced here (the API adds it automatically).
func routesFromModel(ctx context.Context, m *resource_route_table.RouteTableModel, diags *diag.Diagnostics) []api.Route {
	if m.Routes.IsNull() || m.Routes.IsUnknown() {
		return nil
	}
	swap := make([]resource_route_table.RoutesValue, 0, len(m.Routes.Elements()))
	diags.Append(m.Routes.ElementsAs(ctx, &swap, false)...)

	routes := make([]api.Route, len(swap))
	for i := range swap {
		routes[i] = api.Route{
			DestinationIpRange: swap[i].DestinationIpRange.ValueStringPointer(),
			GatewayId:          utils.FromTfStringToStringPtrExcludeEmpty(swap[i].GatewayId),
			NatGatewayId:       utils.FromTfStringToStringPtrExcludeEmpty(swap[i].NatGatewayId),
			VpcPeeringId:       utils.FromTfStringToStringPtrExcludeEmpty(swap[i].VpcPeeringId),
			NicId:              utils.FromTfStringToStringPtrExcludeEmpty(swap[i].NicId),
			VmId:               utils.FromTfStringToStringPtrExcludeEmpty(swap[i].VmId),
		}
	}
	return routes
}

// subnetIDsFromModel collapses the deprecated single subnet_id and the
// subnet_ids set into the desired []string the core link/unlink calls expect.
// subnet_id and subnet_ids are mutually exclusive (enforced by a ConflictsWith
// validator), so at most one is set.
func subnetIDsFromModel(ctx context.Context, m *resource_route_table.RouteTableModel, diags *diag.Diagnostics) []string {
	if !m.SubnetId.IsNull() && m.SubnetId.ValueString() != "" {
		return []string{m.SubnetId.ValueString()}
	}
	if m.SubnetIds.IsNull() || m.SubnetIds.IsUnknown() {
		return nil
	}
	var ids []string
	diags.Append(m.SubnetIds.ElementsAs(ctx, &ids, false)...)
	return ids
}

// linksFromState returns the link ids to unlink on delete (the API deletes a
// route table only once all its subnet links are removed).
func linksFromState(ctx context.Context, m *resource_route_table.RouteTableModel, diags *diag.Diagnostics) []string {
	return utils.TfListToGenericList(func(link resource_route_table.LinkRouteTablesValue) string {
		return link.Id.ValueString()
	}, ctx, m.LinkRouteTables, diags)
}

// serializeNonLocalRoutes serializes the read routes into the routes set,
// excluding the automatic "local" route (which is surfaced separately as
// local_route). The generator cannot partition a single SDK slice into two
// distinct TF attributes, so this is wired via a serialize override.
func serializeNonLocalRoutes(ctx context.Context, apiRoutes *[]api.Route, diags *diag.Diagnostics) types.Set {
	if apiRoutes == nil {
		return types.SetNull(resource_route_table.RoutesValue{}.Type(ctx))
	}
	nonLocal := make([]api.Route, 0, len(*apiRoutes))
	for _, route := range *apiRoutes {
		if route.GatewayId != nil && strings.EqualFold(*route.GatewayId, "local") {
			continue
		}
		nonLocal = append(nonLocal, route)
	}
	return utils.GenericSetToTfSetValue(ctx, serializeRoute, nonLocal, diags)
}

// localRouteFromRoutes extracts the automatic "local" route from the read routes
// into the computed local_route attribute.
func localRouteFromRoutes(ctx context.Context, apiRoutes *[]api.Route, diags *diag.Diagnostics) resource_route_table.LocalRouteValue {
	if apiRoutes != nil {
		for _, route := range *apiRoutes {
			if route.GatewayId != nil && strings.EqualFold(*route.GatewayId, "local") {
				return localRouteValue(ctx, route, diags)
			}
		}
	}
	return resource_route_table.NewLocalRouteValueNull()
}

// subnetIDsFromLinks serializes the subnet links into the subnet_ids set.
// subnet_ids is a set (order-insensitive), so it can be derived directly from
// the read links without consulting prior state.
func subnetIDsFromLinks(ctx context.Context, links *[]api.LinkRouteTable, diags *diag.Diagnostics) types.Set {
	if links == nil {
		return types.SetNull(types.StringType)
	}
	ids := make([]string, 0, len(*links))
	for _, link := range *links {
		if link.SubnetId != nil {
			ids = append(ids, *link.SubnetId)
		}
	}
	set, d := types.SetValueFrom(ctx, types.StringType, ids)
	diags.Append(d...)
	return set
}

func serializeRoute(ctx context.Context, route *api.Route, diags *diag.Diagnostics) resource_route_table.RoutesValue {
	value, d := resource_route_table.NewRoutesValue(
		resource_route_table.RoutesValue{}.AttributeTypes(ctx),
		routeAttributes(*route),
	)
	diags.Append(d...)
	return value
}

func localRouteValue(ctx context.Context, route api.Route, diags *diag.Diagnostics) resource_route_table.LocalRouteValue {
	value, d := resource_route_table.NewLocalRouteValue(
		resource_route_table.LocalRouteValue{}.AttributeTypes(ctx),
		routeAttributes(route),
	)
	diags.Append(d...)
	return value
}

func routeAttributes(route api.Route) map[string]attr.Value {
	return map[string]attr.Value{
		"creation_method":        types.StringPointerValue(route.CreationMethod),
		"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
		"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
		"gateway_id":             types.StringPointerValue(route.GatewayId),
		"nat_gateway_id":         types.StringPointerValue(route.NatGatewayId),
		"vpc_peering_id":         types.StringPointerValue(route.VpcPeeringId),
		"nic_id":                 types.StringPointerValue(route.NicId),
		"state":                  types.StringPointerValue(route.State),
		"vm_id":                  types.StringPointerValue(route.VmId),
	}
}
