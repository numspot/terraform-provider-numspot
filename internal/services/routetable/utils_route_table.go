package routetable

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func RouteTableFromHttpToTf(ctx context.Context, http *numspot.RouteTable) (*RouteTableModel, diag.Diagnostics) {
	var (
		tagsTf     types.List
		diags      diag.Diagnostics
		localRoute RoutesValue
		routes     []numspot.Route
	)

	if http.Routes == nil {
		return nil, diags
	}
	for _, route := range *http.Routes {
		if route.GatewayId != nil && *route.GatewayId == "local" {
			localRoute, diags = routeTableRouteFromAPI(ctx, route)
			if diags.HasError() {
				return nil, diags
			}
		} else {
			routes = append(routes, route)
		}
	}

	// Routes
	tfRoutes, diags := utils.GenericSetToTfSetValue(ctx, RoutesValue{}, routeTableRouteFromAPI, routes)
	if diags.HasError() {
		return nil, diags
	}

	// Links
	tfLinks, diags := utils.GenericListToTfListValue(ctx, LinkRouteTablesValue{}, routeTableLinkFromAPI, *http.LinkRouteTables)
	if diags.HasError() {
		return nil, diags
	}

	// Retrieve Subnet Id:
	var subnetId *string
	for _, assoc := range *http.LinkRouteTables {
		if assoc.SubnetId != nil {
			subnetId = assoc.SubnetId
			break
		}
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	res := RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		LinkRouteTables:                 tfLinks,
		VpcId:                           types.StringPointerValue(http.VpcId),
		RoutePropagatingVirtualGateways: types.ListNull(RoutePropagatingVirtualGatewaysValue{}.Type(ctx)),
		Routes:                          tfRoutes,
		SubnetId:                        types.StringPointerValue(subnetId),
		Tags:                            tagsTf,
		LocalRoute:                      localRoute,
	}

	return &res, diags
}

func routeTableLinkFromAPI(ctx context.Context, link numspot.LinkRouteTable) (LinkRouteTablesValue, diag.Diagnostics) {
	return NewLinkRouteTablesValue(
		LinkRouteTablesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":             types.StringPointerValue(link.Id),
			"main":           types.BoolPointerValue(link.Main),
			"route_table_id": types.StringPointerValue(link.RouteTableId),
			"subnet_id":      types.StringPointerValue(link.SubnetId),
			"vpc_id":         types.StringPointerValue(link.VpcId),
		},
	)
}

func routeTableRouteFromAPI(ctx context.Context, route numspot.Route) (RoutesValue, diag.Diagnostics) {
	return NewRoutesValue(
		RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"creation_method":        types.StringPointerValue(route.CreationMethod),
			"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
			"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
			"gateway_id":             types.StringPointerValue(route.GatewayId),
			"nat_gateway_id":         types.StringPointerValue(route.NatGatewayId),
			"vpc_peering_id":         types.StringPointerValue(route.VpcPeeringId),
			"nic_id":                 types.StringPointerValue(route.NicId),
			"state":                  types.StringPointerValue(route.State),
			"vm_id":                  types.StringPointerValue(route.VmId),
		},
	)
}

func RouteTableFromTfToCreateRequest(tf *RouteTableModel) numspot.CreateRouteTableJSONRequestBody {
	return numspot.CreateRouteTableJSONRequestBody{
		VpcId: tf.VpcId.ValueString(),
	}
}

func RouteTableFromTfToCreateRoutesRequest(route RoutesValue) numspot.CreateRouteJSONRequestBody {
	return numspot.CreateRouteJSONRequestBody{
		DestinationIpRange: route.DestinationIpRange.ValueString(),
		GatewayId:          route.GatewayId.ValueStringPointer(),
		NatGatewayId:       route.NatGatewayId.ValueStringPointer(),
		VpcPeeringId:       route.VpcPeeringId.ValueStringPointer(),
		NicId:              route.NicId.ValueStringPointer(),
		VmId:               route.VmId.ValueStringPointer(),
	}
}

func RouteTableFromTfToDeleteRoutesRequest(route RoutesValue) numspot.DeleteRouteJSONRequestBody {
	return numspot.DeleteRouteJSONRequestBody{
		DestinationIpRange: route.DestinationIpRange.ValueString(),
	}
}

func RouteTablesFromTfToAPIReadParams(ctx context.Context, tf RouteTablesDataSourceModel) numspot.ReadRouteTablesParams {
	return numspot.ReadRouteTablesParams{
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
		LinkRouteTableLinkRouteTableIds: utils.TfStringListToStringPtrList(ctx, tf.LinkRouteTableLinkRouteTableIds),
		LinkSubnetIds:                   utils.TfStringListToStringPtrList(ctx, tf.LinkSubnetIds),
	}
}

func RouteTablesFromHttpToTfDatasource(ctx context.Context, http *numspot.RouteTable) (*RouteTableModelDatasource, diag.Diagnostics) {
	var (
		diags                               diag.Diagnostics
		tagsList                            = types.ListNull(tags.TagsValue{}.Type(ctx))
		linkRouteTablesList                 = types.ListNull(LinkRouteTablesValue{}.Type(ctx))
		routes                              = types.ListNull(RoutesValue{}.Type(ctx))
		routePropagatingVirtualGatewaysList = types.ListNull(RoutePropagatingVirtualGatewaysValue{}.Type(ctx))
	)

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.LinkRouteTables != nil {
		linkRouteTablesList, diags = utils.GenericListToTfListValue(ctx, LinkRouteTablesValue{}, routeTableLinkFromAPIDatasource, *http.LinkRouteTables)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.RoutePropagatingVirtualGateways != nil {
		routePropagatingVirtualGatewaysList, diags = utils.GenericListToTfListValue(
			ctx, RoutePropagatingVirtualGatewaysValue{},
			routeTableRoutePropagatingVirtualGatewaysFromAPIDatasource,
			*http.RoutePropagatingVirtualGateways)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Routes != nil {
		routes, diags = utils.GenericListToTfListValue(ctx, RoutesValue{}, routeFromAPIDatasource, *http.Routes)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &RouteTableModelDatasource{
		Id:                              types.StringPointerValue(http.Id),
		Tags:                            tagsList,
		LinkRouteTables:                 linkRouteTablesList,
		RoutePropagatingVirtualGateways: routePropagatingVirtualGatewaysList,
		VpcId:                           types.StringPointerValue(http.VpcId),
		Routes:                          routes,
	}, nil
}

func routeFromAPIDatasource(ctx context.Context, route numspot.Route) (RoutesValue, diag.Diagnostics) {
	return NewRoutesValue(
		RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"creation_method":        types.StringPointerValue(route.CreationMethod),
			"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
			"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
			"gateway_id":             types.StringPointerValue(route.GatewayId),
			"nat_gateway_id":         types.StringPointerValue(route.NatGatewayId),
			"nic_id":                 types.StringPointerValue(route.NicId),
			"state":                  types.StringPointerValue(route.State),
			"vm_id":                  types.StringPointerValue(route.VmId),
			"vpc_peering_id":         types.StringPointerValue(route.VpcPeeringId),
		},
	)
}

func routeTableLinkFromAPIDatasource(ctx context.Context, link numspot.LinkRouteTable) (LinkRouteTablesValue, diag.Diagnostics) {
	return NewLinkRouteTablesValue(
		LinkRouteTablesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":             types.StringPointerValue(link.Id),
			"main":           types.BoolPointerValue(link.Main),
			"route_table_id": types.StringPointerValue(link.RouteTableId),
			"subnet_id":      types.StringPointerValue(link.SubnetId),
			"vpc_id":         types.StringPointerValue(link.VpcId),
		},
	)
}

func routeTableRoutePropagatingVirtualGatewaysFromAPIDatasource(ctx context.Context, route numspot.RoutePropagatingVirtualGateway) (RoutePropagatingVirtualGatewaysValue, diag.Diagnostics) {
	return NewRoutePropagatingVirtualGatewaysValue(
		RoutePropagatingVirtualGatewaysValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"virtual_gateway_id": types.StringPointerValue(route.VirtualGatewayId),
		},
	)
}
