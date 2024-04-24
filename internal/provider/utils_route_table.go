package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_route_table"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_route_table"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func RouteTableFromHttpToTf(ctx context.Context, http *iaas.RouteTable, defaultRouteDestination string, subnetId *string) (*resource_route_table.RouteTableModel, diag.Diagnostics) {
	// Routes
	var routes []iaas.Route
	if len(*http.Routes) > 0 {
		// Remove "defaulted" route to prevent inconsistent state
		routes = make([]iaas.Route, 0, len(*http.Routes)-1)
		for _, e := range *http.Routes {
			if *e.DestinationIpRange != defaultRouteDestination {
				routes = append(routes, e)
			}
		}
	} else {
		routes = *http.Routes
	}

	tfRoutes, diagnostics := utils.GenericListToTfListValue(ctx, resource_route_table.RoutesValue{}, routeTableRouteFromAPI, routes)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Links
	tfLinks, diagnostics := utils.GenericListToTfListValue(ctx, resource_route_table.LinkRouteTablesValue{}, routeTableLinkFromAPI, *http.LinkRouteTables)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	res := resource_route_table.RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		LinkRouteTables:                 tfLinks,
		VpcId:                           types.StringPointerValue(http.VpcId),
		RoutePropagatingVirtualGateways: types.ListNull(resource_route_table.RoutePropagatingVirtualGatewaysValue{}.Type(ctx)),
		Routes:                          tfRoutes,
		SubnetId:                        types.StringPointerValue(subnetId),
	}

	return &res, nil
}

func routeTableLinkFromAPI(ctx context.Context, link iaas.LinkRouteTable) (resource_route_table.LinkRouteTablesValue, diag.Diagnostics) {
	return resource_route_table.NewLinkRouteTablesValue(
		resource_route_table.LinkRouteTablesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":             types.StringPointerValue(link.Id),
			"main":           types.BoolPointerValue(link.Main),
			"route_table_id": types.StringPointerValue(link.RouteTableId),
			"subnet_id":      types.StringPointerValue(link.SubnetId),
			"vpc_id":         types.StringPointerValue(link.VpcId),
		},
	)
}

func routeTableRouteFromAPI(ctx context.Context, route iaas.Route) (resource_route_table.RoutesValue, diag.Diagnostics) {
	return resource_route_table.NewRoutesValue(
		resource_route_table.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"creation_method":        types.StringPointerValue(route.CreationMethod),
			"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
			"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
			"gateway_id":             types.StringPointerValue(route.GatewayId),
			"nat_gateway_id":         types.StringPointerValue(route.NatGatewayId),
			"vpc_access_point_id":    types.StringPointerValue(route.VpcAccessPointId),
			"vpc_peering_id":         types.StringPointerValue(route.VpcPeeringId),
			"nic_id":                 types.StringPointerValue(route.NicId),
			"state":                  types.StringPointerValue(route.State),
			"vm_id":                  types.StringPointerValue(route.VmId),
		},
	)
}

func RouteTableFromTfToCreateRequest(tf *resource_route_table.RouteTableModel) iaas.CreateRouteTableJSONRequestBody {
	return iaas.CreateRouteTableJSONRequestBody{
		VpcId: tf.VpcId.ValueString(),
	}
}

func RouteTablesFromTfToAPIReadParams(ctx context.Context, tf RouteTablesDataSourceModel) iaas.ReadRouteTablesParams {
	return iaas.ReadRouteTablesParams{
		TagKeys:                         utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:                       utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:                            utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:                             utils.TfStringListToStringPtrList(ctx, tf.Ids),
		RouteVpcPeeringIds:              utils.TfStringListToStringPtrList(ctx, tf.RouteVpcPeeringIds),
		RouteNatGatewayIds:              utils.TfStringListToStringPtrList(ctx, tf.RouteNatGatewayIds),
		RouteVmIds:                      utils.TfStringListToStringPtrList(ctx, tf.RouteVmIds),
		RouteCreationMethods:            utils.TfStringListToStringPtrList(ctx, tf.RouteCreationMethods),
		RouteDestinationIpRanges:        utils.TfStringListToStringPtrList(ctx, tf.RouteDestinationIpRanges),
		RouteDestinationServiceIds:      utils.TfStringListToStringPtrList(ctx, tf.RouteDestinationServiceIds),
		RouteGatewayIds:                 utils.TfStringListToStringPtrList(ctx, tf.RouteGatewayIds),
		RouteStates:                     utils.TfStringListToStringPtrList(ctx, tf.RouteStates),
		VpcIds:                          utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		LinkRouteTableIds:               utils.TfStringListToStringPtrList(ctx, tf.LinkRouteTableIds),
		LinkRouteTableMain:              utils.FromTfBoolToBoolPtr(tf.LinkRouteTableMain),
		LinkRouteTableLinkRouteTableIds: utils.TfStringListToStringPtrList(ctx, tf.LinkRouteTableRouteTableIds),
		LinkSubnetIds:                   utils.TfStringListToStringPtrList(ctx, tf.LinkRouteTableSubnetIds),
	}
}

func RouteTablesFromHttpToTfDatasource(ctx context.Context, http *iaas.RouteTable) (*datasource_route_table.RouteTableModel, diag.Diagnostics) {
	var (
		diags                               diag.Diagnostics
		tagsList                            = types.ListNull(tags.TagsValue{}.Type(ctx))
		linkRouteTablesList                 = types.ListNull(datasource_route_table.LinkRouteTablesValue{}.Type(ctx))
		routePropagatingVirtualGatewaysList = types.ListNull(datasource_route_table.RoutePropagatingVirtualGatewaysValue{}.Type(ctx))
		routesList                          = types.ListNull(datasource_route_table.RoutesValue{}.Type(ctx))
	)

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.LinkRouteTables != nil {
		linkRouteTablesList, diags = utils.GenericListToTfListValue(ctx, datasource_route_table.LinkRouteTablesValue{}, routeTableLinkFromAPIDatasource, *http.LinkRouteTables)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.RoutePropagatingVirtualGateways != nil {
		routePropagatingVirtualGatewaysList, diags = utils.GenericListToTfListValue(
			ctx, datasource_route_table.RoutePropagatingVirtualGatewaysValue{},
			routeTableRoutePropagatingVirtualGatewaysFromAPIDatasource,
			*http.RoutePropagatingVirtualGateways)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Routes != nil {
		routesList, diags = utils.GenericListToTfListValue(ctx, datasource_route_table.RoutesValue{}, routeTableRouteFromAPIDatasource, *http.Routes)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &datasource_route_table.RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		Tags:                            tagsList,
		LinkRouteTables:                 linkRouteTablesList,
		RoutePropagatingVirtualGateways: routePropagatingVirtualGatewaysList,
		Routes:                          routesList,
		VpcId:                           types.StringPointerValue(http.VpcId),
	}, nil
}

func routeTableLinkFromAPIDatasource(ctx context.Context, link iaas.LinkRouteTable) (datasource_route_table.LinkRouteTablesValue, diag.Diagnostics) {
	return datasource_route_table.NewLinkRouteTablesValue(
		datasource_route_table.LinkRouteTablesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":             types.StringPointerValue(link.Id),
			"main":           types.BoolPointerValue(link.Main),
			"route_table_id": types.StringPointerValue(link.RouteTableId),
			"subnet_id":      types.StringPointerValue(link.SubnetId),
			"vpc_id":         types.StringPointerValue(link.VpcId),
		},
	)
}

func routeTableRouteFromAPIDatasource(ctx context.Context, route iaas.Route) (datasource_route_table.RoutesValue, diag.Diagnostics) {
	return datasource_route_table.NewRoutesValue(
		datasource_route_table.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"creation_method":        types.StringPointerValue(route.CreationMethod),
			"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
			"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
			"gateway_id":             types.StringPointerValue(route.GatewayId),
			"nat_gateway_id":         types.StringPointerValue(route.NatGatewayId),
			"vpc_access_point_id":    types.StringPointerValue(route.VpcAccessPointId),
			"vpc_peering_id":         types.StringPointerValue(route.VpcPeeringId),
			"nic_id":                 types.StringPointerValue(route.NicId),
			"state":                  types.StringPointerValue(route.State),
			"vm_id":                  types.StringPointerValue(route.VmId),
			"vm_account_id":          types.StringPointerValue(route.VmAccountId),
		},
	)
}

func routeTableRoutePropagatingVirtualGatewaysFromAPIDatasource(ctx context.Context, route iaas.RoutePropagatingVirtualGateway) (datasource_route_table.RoutePropagatingVirtualGatewaysValue, diag.Diagnostics) {
	return datasource_route_table.NewRoutePropagatingVirtualGatewaysValue(
		datasource_route_table.RoutePropagatingVirtualGatewaysValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"virtual_gateway_id": types.StringPointerValue(route.VirtualGatewayId),
		},
	)
}
