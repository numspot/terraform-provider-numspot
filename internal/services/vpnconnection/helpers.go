package vpnconnection

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/vpnconnection/resource_vpn_connection"
	"terraform-provider-numspot/internal/utils"
)

const routeStateAttr = "state"

// serializeActiveRoutes maps the API routes onto the TF set, dropping routes the
// API reports as already deleted (a transient state we never want in the plan).
// Referenced by the generated serializer via a mapping override.
func serializeActiveRoutes(ctx context.Context, httpRoutes []api.RouteLight, diags *diag.Diagnostics) basetypes.SetValue {
	active := slices.DeleteFunc(slices.Clone(httpRoutes), func(r api.RouteLight) bool {
		return r.State == "deleted"
	})
	return utils.GenericSetToTfSetValue(ctx, serializeRouteValue, active, diags)
}

func serializeRouteValue(ctx context.Context, elt *api.RouteLight, diags *diag.Diagnostics) resource_vpn_connection.RoutesValue {
	value, diagnostics := resource_vpn_connection.NewRoutesValue(
		resource_vpn_connection.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"destination_ip_range": types.StringValue(elt.DestinationIpRange),
			"route_type":           types.StringValue(elt.RouteType),
			routeStateAttr:         types.StringValue(elt.State),
		})
	diags.Append(diagnostics...)
	return value
}

// routesFromPlan builds the static routes passed to the create call (the API
// adds them via follow-up calls rather than the create body).
func routesFromPlan(ctx context.Context, plan *resource_vpn_connection.VpnConnectionModel, diags *diag.Diagnostics) []api.CreateVPNConnectionRoute {
	return deserializeCreateRoutes(routesSetToRoutesSlice(ctx, plan.Routes, diags))
}

// routesToCreate / routesToDelete reconcile the routes set on update: only the
// difference is created/deleted (routes is the sole non-ForceNew attribute).
func routesToCreate(
	ctx context.Context,
	state, plan *resource_vpn_connection.VpnConnectionModel,
	diags *diag.Diagnostics,
) []api.CreateVPNConnectionRoute {
	toCreate, _ := routeDiff(ctx, state, plan, diags)
	return deserializeCreateRoutes(toCreate)
}

func routesToDelete(
	ctx context.Context,
	state, plan *resource_vpn_connection.VpnConnectionModel,
	diags *diag.Diagnostics,
) []api.DeleteVPNConnectionRoute {
	_, toDelete := routeDiff(ctx, state, plan, diags)
	return deserializeDeleteRoutes(toDelete)
}

func routeDiff(
	ctx context.Context,
	state, plan *resource_vpn_connection.VpnConnectionModel,
	diags *diag.Diagnostics,
) (toCreate, toDelete []resource_vpn_connection.RoutesValue) {
	stateRoutes := routesSetToRoutesSlice(ctx, state.Routes, diags)
	planRoutes := routesSetToRoutesSlice(ctx, plan.Routes, diags)
	return utils.Diff(stateRoutes, planRoutes)
}

func routesSetToRoutesSlice(ctx context.Context, list types.Set, diags *diag.Diagnostics) []resource_vpn_connection.RoutesValue {
	return utils.TfSetToGenericList(func(a resource_vpn_connection.RoutesValue) resource_vpn_connection.RoutesValue {
		return resource_vpn_connection.RoutesValue{
			DestinationIpRange: a.DestinationIpRange,
		}
	}, ctx, list, diags)
}

func deserializeCreateRoutes(tfRoutes []resource_vpn_connection.RoutesValue) []api.CreateVPNConnectionRoute {
	routes := make([]api.CreateVPNConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = api.CreateVPNConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}
	return routes
}

func deserializeDeleteRoutes(tfRoutes []resource_vpn_connection.RoutesValue) []api.DeleteVPNConnectionRoute {
	routes := make([]api.DeleteVPNConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = api.DeleteVPNConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}
	return routes
}
