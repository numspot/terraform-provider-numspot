package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_route_table"
)

func RouteTableFromTfToHttp(tf *resource_route_table.RouteTableModel) *api.RouteTableSchema {
	return &api.RouteTableSchema{
		Id:                              tf.Id.ValueStringPointer(),
		NetId:                           tf.NetId.ValueStringPointer(),
		RoutePropagatingVirtualGateways: nil,
		Routes:                          nil,
	}
}

func RouteTableFromHttpToTf(ctx context.Context, http *api.RouteTableSchema, defaultRouteDestination string, subnetId *string) (*resource_route_table.RouteTableModel, diag.Diagnostics) {
	// Routes
	routes := []resource_route_table.RoutesValue{}
	for _, route := range *http.Routes {
		if *route.DestinationIpRange != defaultRouteDestination {
			nroutev, diagnostics := routeTableRouteFromAPI(ctx, &route)
			if diagnostics.HasError() {
				return nil, diagnostics
			}

			routes = append(routes, nroutev)
		}
	}
	tfRoutes, diagnostics := types.ListValueFrom(
		ctx,
		resource_route_table.RoutesValue{}.Type(ctx),
		routes,
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Links
	links := make([]resource_route_table.LinkRouteTablesValue, 0, len(*http.LinkRouteTables))
	for _, link := range *http.LinkRouteTables {
		nlink, diagnostics := routeTableLinkFromAPI(ctx, link)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
		links = append(links, nlink)
	}
	tfLinks, nDiag := types.ListValueFrom(
		ctx,
		resource_route_table.LinkRouteTablesValue{}.Type(ctx),
		links,
	)
	diagnostics.Append(nDiag...)
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

func routeTableRouteFromAPI(ctx context.Context, route *api.RouteSchema) (resource_route_table.RoutesValue, diag.Diagnostics) {
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
