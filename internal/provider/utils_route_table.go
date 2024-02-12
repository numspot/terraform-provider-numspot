package provider

import (
	"context"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_route_table"
)

func RouteTableFromHttpToTf(ctx context.Context, http *api.RouteTableSchema, defaultRouteDestination string, subnetId *string) (*resource_route_table.RouteTableModel, diag.Diagnostics) {
	// Routes
	var routes []api.RouteSchema
	if len(*http.Routes) > 0 {
		// Remove "defaulted" route to prevent inconsistent state
		routes = make([]api.RouteSchema, 0, len(*http.Routes)-1)
		for _, e := range *http.Routes {
			if *e.DestinationIpRange != defaultRouteDestination {
				routes = append(routes, e)
			}
		}
	} else {
		routes = *http.Routes
	}

	tfRoutes, diagnostics := utils.GenericListToTfListValue(ctx, routeTableRouteFromAPI, routes)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Links
	tfLinks, diagnostics := utils.GenericListToTfListValue(ctx, routeTableLinkFromAPI, *http.LinkRouteTables)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	res := resource_route_table.RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		LinkRouteTables:                 tfLinks,
		NetId:                           types.StringPointerValue(http.NetId),
		RoutePropagatingVirtualGateways: types.ListNull(resource_route_table.RoutePropagatingVirtualGatewaysValue{}.Type(ctx)),
		Routes:                          tfRoutes,
		SubnetId:                        types.StringPointerValue(subnetId),
	}

	return &res, nil
}

func routeTableLinkFromAPI(ctx context.Context, link api.LinkRouteTableSchema) (resource_route_table.LinkRouteTablesValue, diag.Diagnostics) {
	return resource_route_table.NewLinkRouteTablesValue(
		resource_route_table.LinkRouteTablesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":             types.StringPointerValue(link.Id),
			"main":           types.BoolPointerValue(link.Main),
			"route_table_id": types.StringPointerValue(link.RouteTableId),
			"subnet_id":      types.StringPointerValue(link.SubnetId),
		},
	)
}

func routeTableRouteFromAPI(ctx context.Context, route api.RouteSchema) (resource_route_table.RoutesValue, diag.Diagnostics) {
	return resource_route_table.NewRoutesValue(
		resource_route_table.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"creation_method":        types.StringPointerValue(route.CreationMethod),
			"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
			"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
			"gateway_id":             types.StringPointerValue(route.GatewayId),
			"nat_service_id":         types.StringPointerValue(route.NatServiceId),
			"net_access_point_id":    types.StringPointerValue(route.NetAccessPointId),
			"net_peering_id":         types.StringPointerValue(route.NetPeeringId),
			"nic_id":                 types.StringPointerValue(route.NicId),
			"state":                  types.StringPointerValue(route.State),
			"vm_account_id":          types.StringPointerValue(route.VmAccountId),
			"vm_id":                  types.StringPointerValue(route.VmId),
		},
	)
}

func RouteTableFromTfToCreateRequest(tf *resource_route_table.RouteTableModel) api.CreateRouteTableJSONRequestBody {
	return api.CreateRouteTableJSONRequestBody{
		NetId: tf.NetId.ValueString(),
	}
}
